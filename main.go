package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

type SnapshotRepository struct {
	Type     string                     `json:"type"`
	Settings SnapshotRepositorySettings `json:"settings"`
}

type SnapshotRepositorySettings struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	BasePath  string `json:"base_path"`
	Compress  bool   `json:"compress"`
}

type SnapshotSettings struct {
	Indices            string `json:"indices"`
	IgnoreUnavailable  bool   `json:"ignore_unavailable"`
	IncludeGlobalState bool   `json:"include_global_state"`
}

type SnapshotList struct {
	Snapshots []SnapshotInfo `json:"snapshots"`
}

type SnapshotInfo struct {
	Snapshot          string            `json:"snapshot"`
	UUID              string            `json:"uuid"`
	VersinoID         int64             `json:"version_id"`
	Version           string            `json:"version"`
	Indices           []string          `json:"indices"`
	State             string            `json:"state"`
	StartTime         string            `json:"start_time"`
	StartTimeInMillis int64             `json:"start_time_in_millis"`
	EndTime           string            `json:"end_time"`
	EndTimeInMillis   int64             `json:"end_time_in_millis"`
	DurationInMillis  int64             `json:"duration_in_millis"`
	Failures          []string          `json:"failures"`
	Shards            SnapshotShardInfo `json:"shards"`
}

type SnapshotShardInfo struct {
	Total      int `json:"total"`
	Failed     int `json:"failed"`
	Successful int `json:"successful"`
}

type Options struct {
	ServiceName string
	Env         string
	URL         string
	Indices     string

	RepositoryFormat string
	SnapshotFormat   string

	RetryIntervalStr string
	retryInterval    time.Duration
	MaxRetries       int

	Bucket    string
	Region    string
	AccessKey string
	SecretKey string

	Restore bool

	Date       snapshotDate
	DateFormat string

	IgnoreUnavailable  bool
	IncludeGlobalState bool
}

type snapshotDate time.Time

func (d *snapshotDate) String() string {
	return time.Time(*d).Format(options.DateFormat)
}

func (d *snapshotDate) Set(s string) error {
	t, err := time.Parse(options.DateFormat, s)
	if err != nil {
		return err
	}
	*d = snapshotDate(t)
	return nil
}

func (o *Options) Validate() error {
	if o.ServiceName == "" {
		return errors.New("-service-name is required")
	}

	if o.Env == "" {
		return errors.New("-env is required")
	}

	var err error
	o.retryInterval, err = time.ParseDuration(o.RetryIntervalStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse -retry-interval")
	}

	if o.Bucket == "" {
		return errors.New("-bucket is required")
	}

	if o.Region == "" {
		o.Region = os.Getenv("AWS_REGION")
	}
	if o.Region == "" {
		return errors.New("-region is required")
	}

	if o.AccessKey == "" {
		o.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if o.AccessKey == "" {
		return errors.New("-access-key is required")
	}

	if o.SecretKey == "" {
		o.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if o.SecretKey == "" {
		return errors.New("-secret-key is required")
	}

	return nil
}

func (o *Options) RepositoryName() string {
	return time.Time(o.Date).Format(o.RepositoryFormat)
}

func (o *Options) SnapshotName() string {
	return time.Time(o.Date).Format(o.SnapshotFormat)
}

func (o *Options) RetryInterval() time.Duration {
	return o.retryInterval
}

func (o *Options) SnapshotRepository() SnapshotRepository {
	return SnapshotRepository{
		Type: "s3",
		Settings: SnapshotRepositorySettings{
			Bucket:    o.Bucket,
			Region:    o.Region,
			AccessKey: o.AccessKey,
			SecretKey: o.SecretKey,
			BasePath:  fmt.Sprintf("%s/%s/%s", o.ServiceName, o.Env, o.RepositoryName()),
		},
	}
}

var (
	options = Options{
		Date: snapshotDate(time.Now()),
	}
)

func main() {
	flag.BoolVar(&options.Restore, "restore", false, "restore mode")
	flag.StringVar(&options.ServiceName, "service-name", "", "service name")
	flag.StringVar(&options.Env, "env", "", "env")
	flag.StringVar(&options.URL, "url", "http://localhost:9200", "URL for Elasticsearch")
	flag.StringVar(&options.Indices, "indices", "*,-.*", "target indices")
	flag.StringVar(&options.RepositoryFormat, "repository-format", "200601", "format of repository name")
	flag.StringVar(&options.SnapshotFormat, "snapshot-format", "02", "format of snapshot name")
	flag.StringVar(&options.RetryIntervalStr, "retry-interval", "1m", "retry interval for API request")
	flag.IntVar(&options.MaxRetries, "max-retries", 0, "max retry count for API request")
	flag.StringVar(&options.Bucket, "bucket", "", "s3 bucket to store snapshots")
	flag.StringVar(&options.Region, "region", "", "s3 region")
	flag.StringVar(&options.AccessKey, "access-key", "", "s3 access key")
	flag.StringVar(&options.SecretKey, "secret-key", "", "s3 secret key")
	flag.Var(&options.Date, "date", "date taken snapshot")
	flag.StringVar(&options.DateFormat, "date-format", "20060102", "date format")
	flag.BoolVar(&options.IgnoreUnavailable, "ignore-unavailable", true, "enable ignore_unavailable option")
	flag.BoolVar(&options.IncludeGlobalState, "include-global-state", true, "enable include_global_state option")

	flag.Parse()

	if err := options.Validate(); err != nil {
		log.Fatalf("failed to parse flag: %v", err)
	}

	if err := createRepository(); err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}
	if options.Restore {
		if err := restoreSnapshot(); err != nil {
			log.Fatalf("failed to restore snapshot: %v", err)
		}
	} else {
		if err := createSnapshot(); err != nil {
			log.Fatalf("failed to create snapshot: %v", err)
		}
	}
	log.Println("complete")
}

func createRepository() error {
	requestURL := options.URL + "/_snapshot/" + options.RepositoryName()
	_, _, errs := gorequest.New().
		Retry(options.MaxRetries, options.RetryInterval(), http.StatusGatewayTimeout).
		Put(requestURL).
		Send(options.SnapshotRepository()).
		End()
	if len(errs) > 0 {
		buf := ""
		for _, err := range errs {
			buf += "\t" + err.Error() + "\n"
		}
		return fmt.Errorf("PUT %s:\n%s", requestURL, buf)
	}
	return nil
}

func createSnapshot() error {
	requestURL := fmt.Sprintf("%s/_snapshot/%s/%s", options.URL, options.RepositoryName(), options.SnapshotName())
	_, _, errs := gorequest.New().
		Retry(options.MaxRetries, options.RetryInterval(), http.StatusGatewayTimeout).
		Put(requestURL).
		Send(&SnapshotSettings{
      Indices: options.Indices,
			IgnoreUnavailable:  options.IgnoreUnavailable,
			IncludeGlobalState: options.IncludeGlobalState,
		}).
		End()
	if len(errs) > 0 {
		buf := ""
		for _, err := range errs {
			buf += "\t" + err.Error() + "\n"
		}
		return fmt.Errorf("PUT %s:\n%s", requestURL, buf)
	}

	var (
		list SnapshotList
		info SnapshotInfo
	)
	for i := 0; i <= options.MaxRetries; i++ {
		resp, body, errs := gorequest.New().Get(requestURL).EndStruct(&list)
		if len(errs) > 0 {
			buf := ""
			for _, err := range errs {
				buf += "\t" + err.Error() + "\n"
			}
			return fmt.Errorf("GET %s:\n%s", requestURL, buf)
		}
		if len(list.Snapshots) == 1 {
			info = list.Snapshots[0]
		}
		log.Printf(
			"GET %s:\n\tstatus_code = %d\n\tbody = %s\n\tstate = %s\n\tshards = %#v\n",
			requestURL,
			resp.StatusCode,
			body,
			info.State,
			info.Shards,
		)
		if resp.StatusCode == 200 && info.State == "SUCCESS" {
			break
		}
		if i < options.MaxRetries {
			time.Sleep(options.RetryInterval())
		}
	}
	return nil
}

func restoreSnapshot() error {
	requestURL := fmt.Sprintf("%s/*", options.URL)
	_, _, errs := gorequest.New().
		Retry(options.MaxRetries, options.RetryInterval(), http.StatusGatewayTimeout).
		Delete(requestURL).
		End()
	if len(errs) > 0 {
		buf := ""
		for _, err := range errs {
			buf += "\t" + err.Error() + "\n"
		}
		return fmt.Errorf("DELETE %s:\n%s", requestURL, buf)
	}

	requestURL = fmt.Sprintf("%s/_snapshot/%s/%s/_restore", options.URL, options.RepositoryName(), options.SnapshotName())
	_, _, errs = gorequest.New().
		Retry(options.MaxRetries, options.RetryInterval(), http.StatusGatewayTimeout).
		Post(requestURL).
		Send(&SnapshotSettings{Indices: options.Indices,
			IgnoreUnavailable:  options.IgnoreUnavailable,
			IncludeGlobalState: options.IncludeGlobalState,
		}).
		End()
	if len(errs) > 0 {
		buf := ""
		for _, err := range errs {
			buf += "\t" + err.Error() + "\n"
		}
		return fmt.Errorf("POST %s:\n%s", requestURL, buf)
	}
	return nil
}

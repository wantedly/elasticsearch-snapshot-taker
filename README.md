[![Docker Repository on Quay](https://quay.io/repository/wantedly/elasticsearch-snapshot-taker/status "Docker Repository on Quay")](https://quay.io/repository/wantedly/elasticsearch-snapshot-taker)

# elasticsearch-snapshot-taker
Backup and Restore Elasticsearch cluster with AWS S3

## Usage

```
Usage of /elasticsearch-snapshot-taker:
  -access-key string
    	s3 access key
  -bucket string
    	s3 bucket to store snapshots
  -date value
    	date taken snapshot (default )
  -date-format string
    	date format (default "20060102")
  -env string
    	env
  -ignore-unavailable
    	enable ignore_unavailable option (default true)
  -include-global-state
    	enable include_global_state option (default true)
  -max-retries int
    	max retry count for API request
  -region string
    	s3 region
  -repository-format string
    	format of repository name (default "200601")
  -restore
    	restore mode
  -retry-interval string
    	retry interval for API request (default "1m")
  -secret-key string
    	s3 secret key
  -service-name string
    	service name
  -snapshot-format string
    	format of snapshot name (default "02")
  -url string
    	URL for Elasticsearch (default "http://localhost:9200")
```

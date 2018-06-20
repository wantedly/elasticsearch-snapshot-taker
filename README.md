# elasticsearch-snapshot-taker
Backup and Restore Elasticsearch cluster with AWS S3

## Usage

```
Usage of /snapshot:
  -access-key string
    	s3 access key
  -bucket string
    	s3 bucket to store snapshots
  -env string
    	env
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

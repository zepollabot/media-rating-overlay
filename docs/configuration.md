# Configuration Guide

This guide explains how to configure Media Rating Overlay to suit your needs.

## Configuration Files

The application uses two main configuration files:
- `configs/config.yaml`: Main configuration for media services and overlay settings
- `configs/config.env.prod.yaml`: Environment-specific configuration for performance, logging, and processing settings

## Main Configuration (config.yaml)

This file contains the core settings for media services and overlay customization.

### Plex Server Configuration

```yaml
plex: # REQUIRED for the script to run
  url: "http://172.17.0.1:32400" # use Docker compose IP for host machine, see "docker network inspect bridge"
  token: "your-plex-token"
  enabled: true
  libraries:
   -  name: "Film"  # Name of your Plex library
      enabled: true  # Whether this library is active
      refresh: false  # Whether to refresh this library
      path: "/Multimedia/Film"  # Path to the library
      filters:
        added_at: last_5_months  # Optional: filter by added date
        titles:
           - "harry potter"  # Optional: filter by specific titles
           - "star wars"
        year: 2010  # Optional: filter by year
      overlay:
        type: "frame"  # Options: "frame" or "bar"
        height: 0.08  # Height of the overlay (as a fraction of screen height)
        transparency: 0.8  # Transparency level (0.0 to 1.0)
```

### TMDB Configuration

```yaml
tmdb:
  enabled: true
  api_key: "your-tmdb-api-key"
  language: "it"  # Language code for TMDB API
  region: "it_IT"  # Region code for TMDB API
```

## Environment Configuration (config.env.prod.yaml)

This file contains environment-specific settings for performance, logging, and processing.

### Performance Settings

```yaml
performance:
  max_threads: 6  # Maximum number of concurrent threads
  library_processing_timeout: 600s  # Timeout for the whole application (10m default)
```

### HTTP Client Settings

```yaml
http_client:
  timeout: 30s  # Request timeout in seconds
  max_retries: 3  # Maximum number of retry attempts
```

### Logger Configuration

```yaml
logger:
  log_file_path: "logs/media-rating-overlay.log"
  log_level: "info"  # debug, info, warn, error
  max_size: 100  # Maximum log file size in megabytes
  max_backups: 5  # Number of backup files to keep
  max_age: 30  # Maximum age of log files in days
  compress: true  # Whether to compress old log files
  service_name: "media-rating-overlay"
  service_version: "1.0.0"
  use_json: true  # Whether to use JSON format for logs
  use_stdout: true  # Whether to output logs to stdout
```

### Processor Settings

```yaml
processor:
  item_processor:
    rating_builder:
      timeout: 30s  # Timeout for rating builder operations
  library_processor:
    default_timeout: 600s  # Timeout for single library processing (10m default)
```

## Getting API Keys

### TMDB API Key
1. Visit [TMDB API](https://www.themoviedb.org/settings/api)
2. Create an account
3. Request an API key

### Plex Token

To get your Plex token:
1. Sign in to [Plex Web](https://app.plex.tv)
2. Browse to a library item and view the XML for it
3. Look in the URL and find the token as the X-Plex-Token value

For more information: https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/

## Best Practices

1. **Performance**
   - Adjust `max_threads` based on your system capabilities
   - Set appropriate timeouts for your environment
   - Use appropriate filter settings to limit processing
   - Enable/disable libraries as needed

2. **Logging**
   - Configure appropriate log levels for your environment
   - Set reasonable log rotation policies
   - Monitor log file sizes

3. **Maintenance**
   - Regularly check logs for errors
   - Update configuration as needed
   - Verify library paths are correct

## Troubleshooting Configuration Issues

Common configuration issues and solutions:

1. **API Authentication Failures**
   - Verify TMDB API key is correct
   - Check API service status

2. **Plex Connection Issues**
   - Verify Plex server URL
   - Check Plex token validity

3. **Library Issues**
   - Verify library paths are correct
   - Verify library names are correct
   - Check filter syntax
   - Ensure proper permissions for file access

4. **Performance Issues**
   - Check timeout settings
   - Verify thread count is appropriate
   - Monitor system resources

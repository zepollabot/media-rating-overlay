# Installation Guide

This guide will help you install and set up Media Rating Overlay on your system.

## Prerequisites

- A running Plex Media Server
- Docker (for Docker installation) + Docker Compose (optional)
- Go 1.20+ (for manual installation)
- API keys for the rating services you want to use (IMDB, TMDB, etc.)


## Installation Methods

1. **Clone the Repository**
   
   **HTTPS**
   ```bash
   git clone https://github.com/zepollabot/media-rating-overlay.git
   ```
   **SSH**
   ```bash
   git clone git@github.com:zepollabot/media-rating-overlay.git
   ```

   ```bash
   cd media-rating-overlay
   ```

2. **Setup the configuration**
   ```bash
   make setup-config
   ```

3. **Configure the Application**

   Edit `configs/config.yaml` with your settings about:
   - Plex server details
   - API keys
   - Libraries settings

   Edit `configs/config.env.prod.yaml` with your settings about:
   - Performance
   - Logging preferences

### Docker Compose Installation

1. **Start the Service**

   ```bash
   docker compose up
   ```

### Manual Installation

1. **Install Go**
   Ensure you have Go 1.20 or later installed:
   ```bash
   go version
   ```
2. **Run the Service**
   ```bash
   make run
   ```

## Post-Installation

1. **Verify Connection to Plex**
   - Check the logs for successful connection
   - Verify the service can access your Plex library

2. **Test Rating Overlays**
   - Check a few media items to ensure ratings are being added
   - Verify the overlay appearance matches your preferences

## Troubleshooting

If you encounter any issues during installation:

1. Check the logs:
   ```bash
   docker compose logs -f  # For Docker installation
   tail -f logs/media-rating-overlay.log  # For manual installation
   ```

2. Verify your configuration:
   - Ensure all required API keys are present
   - Check Plex server connection details
   - Verify file permissions

3. Common issues:
   - API key authentication failures
   - Plex server connection issues
   - Permission problems with media files

## Next Steps

After successful installation:
1. Configure your rating preferences
3. Customize the overlay appearance

See the [Configuration Guide](configuration.md) for detailed information on these topics. 
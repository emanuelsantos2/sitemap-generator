# Sitemap Builder

A powerful Go application for generating XML sitemaps, including support for news sitemaps, with flexible storage options and API management.

## 🚀 Features

- Generate regular and news XML sitemaps
- Support for multiple sitemap indexes
- Configurable datasources
- Chunking for large sitemaps
- Local file system or S3 storage options
- JWT authentication for API protection
- Docker support for easy deployment

## 📋 Prerequisites

- Go 1.16+
- Docker (optional)
- AWS account for S3 storage (optional)

## 🛠️ Installation

### Local Setup

Clone the repository:

```sh
git clone https://github.com/yourusername/sitemap-builder.git
cd sitemap-builder
```

Install dependencies:

```sh
go mod tidy
```

Set up environment variables:

Create a `.env` file in the project root:

```sh
JWT_SECRET=your_jwt_secret_here
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
```

Run the application:

```sh
go run main.go
```

### Docker Setup

Build the Docker image:

```sh
docker build -t sitemap-builder .
```

Run the container:

```sh
docker run -d -p 3000:3000 \
  -e JWT_SECRET=your_jwt_secret_here \
  -e AWS_ACCESS_KEY_ID=your_aws_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_aws_secret_key \
  -v ./sitemaps:/app/sitemaps \
  -v ./data:/app/data \
  sitemap-builder
```

## ⚙️ Configuration

The application uses an `init.json` file for initial configuration. Example structure:

```json
{
  "admin": {
    "username": "admin",
    "password": "securepassword123"
  },
  "storage_configs": [
    {
      "mode": "s3",
      "bucket": "my-sitemaps-bucket",
      "region": "us-west-2",
      "endpoint": "s3.amazonaws.com",
      "path": "production/sitemaps/"
    }
  ],
  "datasources": [
    {
      "name": "default-sqlite",
      "type": "sqlite",
      "connection_string": "data/default.db"
    }
  ],
  "sitemap_indexes": [
    {
      "name": "main-sitemap",
      "storage_config_id": 1,
      "sitemaps": [
        {
          "name": "products",
          "config": {
            "datasource": "default-sqlite",
            "table_name": "products",
            "base_url": "example.com",
            "url_pattern": "/{language}/{slug}",
            "change_frequency": "weekly",
            "priority": 0.8
          }
        }
      ]
    }
  ]
}
```

## 🔑 API Endpoints

- `POST /api/auth/login` - Authenticate and receive JWT token
- `GET /api/sitemap-index` - List all sitemap indexes
- `POST /api/sitemap-index` - Create a new sitemap index
- `GET /api/sitemap` - List all sitemaps
- `POST /api/sitemap` - Create a new sitemap
- `POST /api/generate` - Trigger sitemap generation (protected route)

## 📘 Usage

1. Authenticate using the login endpoint to get a JWT token.
2. Use the token in the `Authorization` header for subsequent requests.
3. Create sitemap indexes, sitemaps, and configurations using the API.
4. Trigger sitemap generation using the generate endpoint.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/yourusername/sitemap-builder/issues).

## 📄 License

This project is MIT licensed.

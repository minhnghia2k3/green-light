<div style="display: flex; justify-content: center;">
    <img alt="golang-icon" height="300" src="assets/golang.png" width="300"/>
</div>

# Greenlight

[Greenlight](https://minhnghia2k3-greenlight-f2bc5d52ac1f.herokuapp.com/swagger/index.html) is a **Go application** offers a **reliable, fast, and secure** API for movies management.

## Key features
🔧 Seamlessly perform **CRUD** operations on movie data

🔑 Robust **authentication** and permission-based **authorization**

🧾 Enhanced movie list management with **filtering**, **sorting**, and **pagination**

🏃‍♂️ Efficient concurrency control leveraging **Context**, **Goroutines**, **Defer**, and **Channels**

📪 Enable background email sending

❌ Comprehensive **logging** and **error handling**

🔐 Implement IP-based **rate limiting** for security

🛠 Ensure **graceful application shutdown** and efficient **metrics** collection

🛠 Support for building, versioning, and quality control for continuous improvement

📢 Public API using **Swagger**

## Diagrams

![Database Diagram](assets/greenlight.png)

![Data Protection Diagram](assets/data-protection.png)

## Prerequisites

- [Go programming language](https://go.dev/doc/install)
- [Ubuntu WSL2](https://ubuntu.com/desktop/wsl)
- PostgresDB running on port `:5432` by default.
- [Mailtrap](https://mailtrap.io/) API token

## Installation

    $ git clone https://github.com/minhnghia2k3/green-light.git

## Usage

- Creating `greenlight` database

      postgres=# CREATE DATABASE greenlight
      postgres=# \c greenlight

- Creating `citext` extension

      greenlight=# CREATE EXTENSION citext;

- Copying `.env.example` to `.env` (modify fields if needed)

      $ cp .env.example .env

- Running application
      
      $ make migrate.up
      $ make run

## License

[Alex Edwards's Lets Go Further](https://lets-go-further.alexedwards.net/)

[MIT](https://choosealicense.com/licenses/mit/)


# GoCommerceAPI

## Overview
A REST API focusing on user management, product management, wallet functionality, and transaction processing. The project cover various aspects, including data management, error handling, and simple security such as access and refresh token management. This project is built with golang, postgres for data storage and redis for caching blocked tokens.

## Features
- **Sign Up**: Register new users
- **Login**: Authenticate users and provide access and refresh token.
- **Logout**: Invalidate access token by adding to the blocklist in redis and invalidate refresh token by delete the token from postgres
- **Access & Refresh Token**: Use JWTs for access token and UUID for refresh token for session management.
- **CRUD Product**: create, read, update and delete product.
- **deposit**: add wallet balance.
- **withdrawal**: reduce wallet balance.
- **Purchase Product**: purchase product and pay with user wallet.
- **transfer**: transfer from wallet to wallet

## Technologies Used
- **Programming Language**: Golang
- **Database**: Postgresql
- **Cache**: Redis
- **Token**: JWT
- **Docker**: Postgresql & Redis
- **Gin**: Golang web framework
- **Pgx**: driver for postgresql

## Getting Started

### Instalation
- **Clone the Repository**:
```
git clone https://github.com/dwiw96/GoCommerceAPI.git
cd go/
```
- **Set up Environment Variable**:
```
open .env file
```
- **Install Dependencies**:
```
go mod tidy
```
- **Run Database Migration**
```
open Makefile
make migrate-up
```
- **Start the App**:
```
go run main.go
```
- **Access the API**:
The default would be be in http://localhost:8080

## Documentation
- **/docs/user-api.json**: openAPI spesicifation
- **/docs/db.erd**: database entity relationship
- **/docs/simple-auth-system.postman_collection.json**: postman specification

## License
This `README.md` includes all the key sections for a backend authentication API project, providing an overview of the functionality, setup instructions, API documentation, security considerations, and sequence diagrams. You can adjust details as needed, such as adding links to specific files or diagrams.
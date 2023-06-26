# Job Aggregator
A web tool aggregates job postings from various job banks. 

## About
The tool is designed to simplify job applications, specifically designed to reduce the overhead required to identify relevent jobs through various job banks (such as LinkedIn, Indeed, Glassdoor, etc). 


## Why is this project created?
The project is created for two key reasons: **Learning** and **utility**

### Learning
The project presents an opportunity for me to get familar with Backend Development namely: 
- Familarization in Golang
- Practice building micro-services architecture
- Familarization with Docker/K8s
- Database Setup

### Utility
Internships, co-op, and full time job applications are a natural part of my foreseeable activity, therefore I wanted to create a tool that can help increase my efficiency. 

## User Story
- As a user, I want to agregrate relevant job postings based on
  - Keywords
  - Location
  - Post Availability
  - Seed URLs of job banks


## Helpful Project Tips
### Run Linter:
Go to file directory and run the following command:
`golangci-lint run [dir_name containing .go files]`

### Run Test:
Go to the test file directory and run the following command:
`go test [dir_name containing go_test files]` ex.  `go test ./cmd`

#### Options:
Verbose:
`go test -v`

Code Coverage:
`go test -cover`

## Acknowledgements

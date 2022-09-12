# ocdtracker-api
This repository contains a GoLang web application that exposes APIs over HTTP. The purpose of this backend is to communicate with mobile/web apps and help users who suffer from Obsessive Compulsive Disorder track their recovery process and reduce rumination ([article](https://thepsychologygroup.com/ruminating-thoughts-and-anxiety/)).  

## Technologies
The application is designed to be used with [Amazon Web Services](https://aws.amazon.com/). The following services are in use:
- VPC (Virtual Private Cloud)
- EC2 (Elastic Compute Cloud)
- RDS (Relational Database Service)
- IAM (Identity and Access Management)
- ECR (Elastic Container Registry)
- S3 (Simple Storage Service)
- Secrets Manager
- CloudFormation

## Setup
1. Create a Firebase project and enable `Auth`
2. Generate private key for the [Firebase Admin SDK](https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk) service account in the same project
3. Create an AWS account with billing enabled in order to be able to provision all required cloud resources
4. Create an S3 bucket called `ocdtracker-api` and upload the file from step 2, but rename it to `google-app-creds.json`
5. Download the [CloudFormation template](https://raw.githubusercontent.com/cecobask/ocdtracker-api/main/cloudformation.template) and create a CloudFormation stack with it
6. Wait until all resources in the stack are fully provisioned (~10m)
7. Find the EC2 instance with name `ocdtracker-api` and get the public ip address. This is the base url for the REST API...

## REST API
Authentication is delegated to [Firebase Auth](https://firebase.google.com/docs/auth) because it provides an efficient way for mobile/web apps to communicate with the API. It gives users the freedom to choose from various auth strategies: Google, email/password, phone number.  
In order to gain access to the API, the users have to attach an authorization (bearer) token to each request. All requests to the web server are tied to your personal account. It's not possible to access or modify other users' data.

### /ocdlog
- `GET`: fetch all ocd logs
- `POST`: create a single ocd log entry
- `DELETE`: remove all ocd logs

### /ocdlog/{id}
- `GET`: fetch a single ocd log entry
- `PATCH`: update a single ocd log entry
- `DELETE`: remove a single ocd log entry

### /account/me
- `GET`: fetch account data
- `PATCH`: update account data
- `DELETE`: remove account and its data

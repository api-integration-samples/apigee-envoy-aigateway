# Apigee Envoy AI Gateway Sample
This sample shows how to setup Apigee X with the Envoy proxy to offer AI services to multiple models with unified authentication, authorization, analytics & monitoring.

## Prequisites
To use this sample, you need to setup Gemini and Mistral in Vertex AI. You will also need Golang and Docker installed on your system (Google Cloud Shell includes these tools), as well as the apigeecli Apigee automation tool.

## Deployment
To deploy, follow these steps.
```sh
# Step 0 - first copy the env file and set your PROJECT_ID for your Apigee X & Vertex AI services
cp 0.env.sh 0.env.local.sh
# edit copied local file
nano 0.env.local.sh
# source variable
source 0.env.local.sh

# Step 1 - build the Go filter for envoy that will extract token counts from the payloads
./1.build.sh

# Step 2 - deploy Apigee X assets
./2.deploy.sh


```
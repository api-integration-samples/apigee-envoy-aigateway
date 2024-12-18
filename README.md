# Apigee Envoy AI Gateway Sample
This sample shows how to setup Apigee X with the Envoy proxy to offer AI services to multiple models with unified authentication, authorization, analytics & monitoring.

## Prequisites
To use this sample, you need to setup Gemini and Mistral in Vertex AI. You will also need Golang and Docker installed on your system (Google Cloud Shell includes these tools).

Additional prerequisites:
- [apigeecli](https://github.com/apigee/apigeecli/tree/main) tool.
- [Apigee Envoy Adapter](https://cloud.google.com/apigee/docs/api-platform/envoy-adapter/v2.0.x/concepts) installed and configured.

## Deployment
To deploy, follow these steps.
```sh
# Step 0 - first copy the env file and set your PROJECT_ID for your Apigee X & Vertex AI services
cp 0.env.sh 0.env.local.sh
# edit copied local file
nano 0.env.local.sh
# source variables
source 0.env.local.sh

# Step 1 - build the Go filter for envoy that will extract token counts from the payloads
./1.build.sh

# Step 2 - int GCP & Apigee X assets
./2.init.sh

# Step 3 - run envoy proxy
./3.start.sh
```

## Testing
After running each of these calls, you can create a custom report in Apigee X using the dc_genai data collectors and see the token consumption to each of the models, regardless if from Apigee X or from Envoy with the Apigee Adapter.
```sh
# get an API key and set here
API_KEY=

# call full api with gemini pro
curl -i -X POST "https://34-8-159-97.nip.io/v1/genai/models/gemini" \
	-H "x-api-key: $API_KEY" \
	-H "Content-Type: application/json; charset=utf-8" \
	--data-binary @- << EOF

{
  "prompt": "Write a story about a magic backpack."
}
EOF

# call full api with mistral
curl -i -X POST "https://34-8-159-97.nip.io/v1/genai/models/mistral" \
	-H "x-api-key: $API_KEY" \
	-H "Content-Type: application/json; charset=utf-8" \
	--data-binary @- << EOF

{
  "prompt": "Write a story about a magic backpack."
}
EOF

# call envoy streaming api with gemini flash
curl -i -X POST "http://localhost:10000/gemini15-flash" \
	-H "Host: gemini.googleapis.com" \
	-H "x-api-key: $API_KEY" \
	-H "Content-Type: application/json; charset=utf-8" \
	-H "Accept: text/event-stream" \
	--data-binary @- << EOF

{
  "model": "google/gemini-1.5-flash-002",
  "stream": true,
  "messages": [{
    "role": "user",
    "content": "Write a story about a magic backpack."
  }]
}
EOF

# call envoy api with mistral nemo
curl -i -X POST "http://localhost:10000/mistral2407-nemo" \
	-H "Host: mistral.googleapis.com" \
	-H "x-api-key: $API_KEY" \
	-H "Content-Type: application/json; charset=utf-8" \
	--data-binary @- << EOF

{
	"model": "mistral-nemo",
	"temperature": 0,
	"messages": [
		{
			"role": "user",
			"content": "Write a story about a magic backpack."
		}
	]
}
EOF

# call envoy large custom model
curl -i -X POST "http://localhost:10000/large-payload-model" \
	-H "Host: largemodel" \
	-H "x-api-key: $API_KEY" \
	-H "Content-Type: application/json; charset=utf-8" \
	--data-binary @- << EOF

{
	"model": "largemodel",
	"temperature": 0,
	"messages": [
		{
			"role": "user",
			"content": "Write a story about a magic backpack."
		}
	]
}
EOF
```
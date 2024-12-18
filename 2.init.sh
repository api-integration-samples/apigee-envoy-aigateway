# enable services
gcloud services enable aiplatform.googleapis.com --project $PROJECT_ID
gcloud services enable run.googleapis.com --project $PROJECT_ID
gcloud services enable cloudbuild.googleapis.com --project $PROJECT_ID
sleep 5

# create service account and grant access
gcloud iam service-accounts create genaiservice \
    --description="Service account to manage Gen AI model access" \
    --display-name="GenAIService" --project $PROJECT_ID
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:genaiservice@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/aiplatform.user" --project $PROJECT_ID

# create Apigee data collectors for llm tokens
apigeecli datacollectors create -n dc_genai_model_name -d "Name of the Gen AI model" -p STRING -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_prompt_tokens -d "Gen AI model prompt token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_completion_tokens -d "Gen AI model completion token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_total_tokens -d "Gen AI model total token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)

# deploy Apigee proxies
cd ./src/main/apigee/apiproxies/GenAI-Analytics-v1
apigeecli apis create bundle -f apiproxy --name "GenAI-Analytics-v1" -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli apis deploy -n "GenAI-Analytics-v1" -o $PROJECT_ID -e $APIGEE_ENV -t $(gcloud auth print-access-token) --ovr
cd ../GenAI-Models-v1
apigeecli apis create bundle -f apiproxy --name "GenAI-Models-v1" -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli apis deploy -n "GenAI-Models-v1" -o $PROJECT_ID -e $APIGEE_ENV -s "genaiservice@$PROJECT_ID.iam.gserviceaccount.com" -t $(gcloud auth print-access-token) --ovr
cd ../../../../..

# create Apigee products
cd ./src/main/apigee/products
apigeecli products import -f products.json -o $PROJECT_ID -t $(gcloud auth print-access-token)
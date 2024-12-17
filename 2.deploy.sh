# create Apigee data collectors for llm tokens
apigeecli datacollectors create -n dc_genai_model_name -d "Name of the Gen AI model" -p STRING -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_prompt_tokens -d "Gen AI model prompt token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_completion_tokens -d "Gen AI model completion token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)
apigeecli datacollectors create -n dc_genai_total_tokens -d "Gen AI model total token count" -p INTEGER -o $PROJECT_ID -t $(gcloud auth print-access-token)

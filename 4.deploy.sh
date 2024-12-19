# deploy to cloud run
gcloud run services replace cloud-run-service.yaml --project $PROJECT_ID --region $REGION
# set public access policy
gcloud run services set-iam-policy apigee-envoy-gateway cloud-run-policy.yaml --project $PROJECT_ID --region $REGION
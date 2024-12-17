# allow default credentials to be read by docker process
chmod o+r $HOME/.config/gcloud/application_default_credentials.json

# run in docker with config and credentials access
docker run --rm -it \
--network host \
-v $(pwd)/envoy-config.yaml:/etc/envoy/config.yaml \
-v $(pwd)/filter.so:/etc/envoy/filter.so \
-e GOOGLE_APPLICATION_CREDENTIALS=/etc/envoy/application_default_credentials.json \
-v $HOME/.config/gcloud/application_default_credentials.json:/etc/envoy/application_default_credentials.json:ro \
-p 9901:9901 -p 10000:10000 \
envoyproxy/envoy:contrib-dev -c /etc/envoy/config.yaml
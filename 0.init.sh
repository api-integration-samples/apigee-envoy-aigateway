# This script copies the environment variables file to a local version, and generates a random storage name
cp 0.env.sh 0.env.local.sh
RANDOM_SUFFIX=$(head /dev/urandom | tr -dc a-z0-9 | head -c5)
sed -i "/export BUCKET_NAME=envoyfs-/c\export BUCKET_NAME=envoyfs-$RANDOM_SUFFIX" 0.env.local.sh
sed -i "/export PROJECT_ID=/c\export PROJECT_ID=$1" 0.env.local.sh
#! /bin/sh

# Check for BitBucket BasicAuth credentials
if [[ -z $BITBUCKET_USERNAME ]]; then
    >&2 echo "BITBUCKET_USERNAME env variable is not set"
    exit 1
fi

if [[ -z $BITBUCKET_PASSWORD ]]; then
    >&2 echo "BITBUCKET_PASSWORD env variable is not set"
    exit 1
fi

# Replace environment variables in a file with their actual values
PAYLOAD=`(echo "cat <<EOF" ; cat example/payload/$@.json; echo EOF) | sh`

# Set git checkout destination in current directory
DESTINATION=$PWD/destination


if [ -d $DESTINATION ]; then
    echo "Directory $DESTINATION exists; It may lead to false negative results during tests of the In stage."
fi

echo $PAYLOAD | go run cmd/$@/main.go $DESTINATION
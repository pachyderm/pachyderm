fileLocations="**/generated/**"

if ! git diff --quiet ${fileLocations}
then
  echo "Looks like you forgot to run \"make graphql\"!"
  echo "$(git diff --color-words)"
  exit 1
fi

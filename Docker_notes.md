Running sequoia docker image

export SEED_PHRASE="{\"seed_phrase\":\"forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm\",\"derivation_path\":\"m/44'/118'/0'/0/0\"}"

docker run --env SEED_PHRASE sequoia:latest

Running docker compose
Replace with your seed phrase in .env
then run `docker compose up`

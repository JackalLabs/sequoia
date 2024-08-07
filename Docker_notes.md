# Get started
Docker
- Replace with your seed phrase in .env
- then run `docker compose up`

Setting up Grafana
- Go to http://localhost:3000
- u:admin p:admin
- Connections > Add new connections
- Search 'Prometheus' > Select it
- Click 'Add new data source' (blue button)
- Connection | Prometheus server URL *: `http://prometheus:9090`
- Click 'Save & test'

Setting up Dashboard
- Dashboards > New (dropdown) > Import
- Select `example_sequoia_dashboard.json`
- Select the Data Source you added earlier
- Click Import







## Running sequoia docker image (for testing docker img)

export SEED_PHRASE="{\"seed_phrase\":\"forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm\",\"derivation_path\":\"m/44'/118'/0'/0/0\"}"

docker run -p 3334:3334 --env SEED_PHRASE sequoia:latest
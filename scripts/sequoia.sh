sequoia init 

# yq -i '.proof_interval=120' $PROV_HOME/config.yaml
# yq -i '.queue_interval=7' $PROV_HOME/config.yaml
# yq -i '.chain_config.rpc_addr="http://localhost:26657"' $PROV_HOME/config.yaml
# yq -i '.chain_config.grpc_addr="localhost:9090"' $PROV_HOME/config.yaml
# yq -i '.domain="http://localhost:3334"' $PROV_HOME/config.yaml
# yq -i '.data_directory="'$PROV_HOME'/data"' $PROV_HOME/config.yaml
# yq -i '.api_config.port=3334' $PROV_HOME/config.yaml
#
rm /root/.sequoia/provider_wallet.json

echo "{\"seed_phrase\":\"forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm\",\"derivation_path\":\"m/44'/118'/0'/0/0\"}" > /root/.sequoia/provider_wallet.json

sequoia start

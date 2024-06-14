make install

sequoia init 

yq -i '.proof_interval=120' /root/.sequoia/config.yaml -y
yq -i '.queue_interval=7' /root/.sequoia/config.yaml -y
yq -i '.chain_config.rpc_addr="https://jackal-t-rpc.noders.services:443"' /root/.sequoia/config.yaml -y
yq -i '.chain_config.grpc_addr="jkl.grpc.t2.stavr.tech:5913"' /root/.sequoia/config.yaml -y
yq -i '.domain="http://localhost:3334"' /root/.sequoia/config.yaml -y
yq -i '.data_directory="/root/.sequoia/data"' /root/.sequoia/config.yaml -y
yq -i '.api_config.port=3334' /root/.sequoia/config.yaml -y

rm /root/.sequoia/provider_wallet.json

echo "{\"seed_phrase\":\"forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm\",\"derivation_path\":\"m/44'/118'/0'/0/0\"}" > /root/.sequoia/provider_wallet.json

sequoia start

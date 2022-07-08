
GeoJSON vector server and web client to use NSW Fuel API

NSW Fuel API:
https://api.nsw.gov.au/DeveloperApp

Geocoding API:
https://mappify.io/

Example run in dev mode (HTTP):
```sh
./fuelnearme --config=app.yml server --dev-mode --port=8010
```

Example run with TLS:
```sh
/fuelnearme --config=app.yml server --port=8010 --tls-cert=service.cert --tls-cert-key=service.key
```


# teslatrack

## Local dev

```bash
docker-compose-up
```

## Hosted on DigitalOcean

```bash
redis-cli -h 157.230.142.205
auth FkNU6btkbjp+RwIG9529yJZG+EfNboVHEC6FzhpifbNMC0fIPC/MJP0/kvo3GYuT7LgkhGDVfE1gEDch
```

Firewall Rules:

* inbound TCP 22 from 24.5.125.231
* inbound TCP 6379 from 24.5.125.231

DigitalOcean Services:
* Managaed Databases - Postgres
* Droplets - Redis
* Kubernetes
* Firewalls
* Load Balancers
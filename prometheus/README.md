To deploy simply run:

```bash
hack/cluster-monitoring/deploy
```

After all pods are ready, you can reach:

* Prometheus UI on node port `30900`
* Alertmanager UI on node port `30903`
* Grafana on node port `30902`

* Also check grafana on `https://APISERVERIP:6443/api/v1/namespaces/monitoring/services/grafana/proxy`

To tear it all down again, run:

```bash
hack/cluster-monitoring/teardown
```

## To deploy the coin exporter
* kubectl create -f coin-exporterk8s.yaml

## After coin exporter deployed and working, goto grafana dashboard, login admin/admin and import the coinmarketcap-grafanadashboard.json file to create the dashboard (choose Prometheus as datasource)

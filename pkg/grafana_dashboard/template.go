package grafana_dashboard

const grafanaTemplate = `{
  "id": 1,
  "slug": "",
  "title": "{{.Title}}",
  "originalTitle": "",
  "tags": null,
  "style": "dark",
  "timezone": "browser",
  "editable": true,
  "hideControls": false,
  "sharedCrosshair": false,
  "panels": null,
  "rows": {{.Indicators}},
  "templating": {
    "list": null
  },
  "annotations": {
    "list": null
  },
  "schemaVersion": 0,
  "version": 0,
  "links": null,
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {
    "now": true,
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  }
}`

const grafanaIndicatorTemplate = `{
  "title": "{{.Title}}",
  "showTitle": false,
  "collapse": false,
  "editable": true,
  "height": "250px",
  "panels": [
    {
      "editable": false,
      "error": false,
      "gridPos": {},
      "id": 1,
      "isNew": true,
      "renderer": "flot",
      "span": 12,
      "title": "{{.Title}}",
      "transparent": false,
      "type": "graph",
      "aliasColors": {},
      "bars": false,
      "fill": 0,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": false,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 2,
      "nullPointMode": "connected",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "refId": "",
          "expr": "{{.Promql}}",
          "intervalFactor": 1,
          "format": "time_series"
        }
      ],
      "tooltip": {
        "shared": true,
        "value_type": "individual"
      },
      "x-axis": true,
      "y-axis": true,
      "xaxis": {
        "format": "",
        "logBase": 0,
        "show": false
      },
      "yaxes": [
        {
          "format": "short",
          "logBase": 1,
          "show": true
        },
        {
          "format": "short",
          "logBase": 1,
          "show": true
        }
      ],
      "thresholds": {{.Thresholds}}
    }
  ],
  "repeat": null
}`

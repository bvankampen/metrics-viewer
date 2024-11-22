package config

const DEFAULT_CONFIG = `settings:
  scrape_interval: 1
metrics:
  - apiserver_flowcontrol_rejected_requests_total
  - apiserver_flowcontrol_current_inqueue_requests
  - apiserver_flowcontrol_request_wait_duration_seconds
  - apiserver_flowcontrol_current_limit_seats
  - apiserver_flowcontrol_lower_limit_seats
  - apiserver_flowcontrol_upper_limit_seats
  - apiserver_flowcontrol_nominal_limit_seats
`

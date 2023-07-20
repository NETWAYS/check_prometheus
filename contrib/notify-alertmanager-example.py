#!/usr/bin/python


from argparse import ArgumentParser
from urllib import request
import json
import socket
import sys


DEBUG = False


def cli(args):
  """
  Commmand Line Arguments
  """

  parser = ArgumentParser(description='Send notifications to the Prometheus Alertmanager')

  parser.add_argument('--hostname', type=str, required=True)
  parser.add_argument('--service', type=str, default='hostalive', required=True)
  parser.add_argument('--output', type=str, default='')
  parser.add_argument('--state', type=int, required=True, choices=[0,1,2,3])
  parser.add_argument('--alert-api-url', type=str, default='http://localhost:9093/api/v1/alerts')

  parser.add_argument('--debug', action='store_true')
  parser.set_defaults(debug=False)

  return parser.parse_args(args)


def post_alert(url, data):
  """
  HTTP Post the data to URL
  """
  if DEBUG:
    print('DEBUG: Posting Alert')

  content = None
  try:
    req = request.Request(url, method="POST")
    req.add_header('Content-Type', 'application/json')
    r = request.urlopen(req, data=data)
    content = r.read()
  except Exception as e:
    print('ERROR:', e)
    sys.exit(1)

  if DEBUG:
    print('DEBUG: Server Return:')
    print(content)


def generate_alert(args):
  if DEBUG:
    print('DEBUG: Generating Alert')

  # Default Value
  state_name_mapping = {
    "0": "Up",
    "1": "Up",
    "2": "Down",
    "3": "Down"
  }

  # Change mapping if hostalive
  if args.service != "hostalive":
    state_name_mapping = {
      "0": "OK",
      "1": "Warning",
      "2": "Critical",
      "3": "Unknown"
    }

  alert_state_name_mapping = {
    "0": "resolved",
    "1": "firing",
    "2": "firing",
    "3": "firing"
  }

  status = state_name_mapping[str(args.state)]
  alert_status = alert_state_name_mapping[str(args.state)]

  return [{
    "status": alert_status,
    "generatorURL": "foo",
    "labels": {
      "alertname": f"{args.service}_{args.hostname}",
      "instance": args.hostname,
      "service": args.service,
    },
    "annotations": {
      "summary": f"Service {args.service} on {args.hostname} is {status}",
    }
  }]

def main(args):

  DEBUG = args.debug

  if DEBUG:
    print('DEBUG: CLI Arguments')
    print('DEBUG:', args)

  alert = generate_alert(args)
  alert_json = json.dumps(alert).encode()

  if DEBUG:
    print('DEBUG: Alert to be posted')
    print('DEBUG:', alert)

  post_alert(args.alert_api_url, alert_json)

if __name__ == "__main__":
  ARGS = cli(sys.argv[1:])
  main(ARGS)

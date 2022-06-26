import os
import collections
import hashlib

import dotenv
from influxdb_client import InfluxDBClient


def get_sessions(query_api, nid='79796f3e-d63d-4608-9ce8-f375eb91129f', inactivity_threshold=60*30):
    query = f' from(bucket:"{os.getenv("INFLUXDB_BUCKET")}")\
        |> range(start: -10d)\
        |> filter(fn:(r) => r["event_type"] == "page_load" or r["event_type"] == "page_click")\
        |> filter(fn: (r) => r["website_id"] == "{nid}")\
        |> filter(fn: (r) => r["_field"] == "user_agent" or r["_field"] == "cpu_cores" or r["_field"] == "device_memory" or r["_field"] == "gpu")\
        |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")'

    visitors = collections.defaultdict(list)
    tables = query_api.query(org=os.getenv('INFLUXDB_ORGANIZATION'), query=query)
    for table in tables:
        for record in table.records:
            visitor_id = ''
            for k in ['user_agent', 'cpu_cores', 'device_mempory', 'gpu']:
                try:
                    visitor_id += str(record[k])
                except KeyError:
                    continue
            hash_object = hashlib.sha512(visitor_id.encode('utf-8'))
            digest = hash_object.hexdigest()
            visitors[digest].append(record['_time'])

    sessions = list()
    for visitor_id in visitors:
        start = None
        previous = None
        for t in sorted(visitors[visitor_id]):
            if previous is None:
                start = previous = t
            else:
                d = (t - previous).total_seconds()
                if d > inactivity_threshold:
                    sessions.append([start, t, visitor_id])
                    previous = None
                else:
                    previous = t
        if previous is not None:
            sessions.append([start, previous, visitor_id])

    return sessions


if __name__ == '__main__':
    dotenv.load_dotenv(os.path.join('..', '..', 'data', 'input', 'credentials.env'))
    client = InfluxDBClient(url=os.getenv('INFLUXDB_URL'), token=os.getenv('INFLUXDB_TOKEN'), org=os.getenv('INFLUXDB_ORGANIZATION'))
    query_api = client.query_api()

    sessions = get_sessions(query_api)    
    for i, (start, end, visitor_id) in enumerate(sessions):
        print(i, start, end, visitor_id)

    
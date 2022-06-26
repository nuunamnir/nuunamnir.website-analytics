import os
import collections
import hashlib
import datetime
import calendar

import dotenv
from influxdb_client import InfluxDBClient
import pandas
import matplotlib.pyplot
import seaborn


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


def dates_by_year(year):
    for month in range(1, 13):
        for day in range(1, calendar.monthrange(year, month)[1] + 1):
            yield datetime.date(year, month, day)


if __name__ == '__main__':
    dotenv.load_dotenv(os.path.join('..', '..', 'data', 'input', 'credentials.env'))
    client = InfluxDBClient(url=os.getenv('INFLUXDB_URL'), token=os.getenv('INFLUXDB_TOKEN'), org=os.getenv('INFLUXDB_ORGANIZATION'))
    query_api = client.query_api()

    sessions = get_sessions(query_api)    

    sessions_by_year = collections.defaultdict(int)
    sessions_by_day = collections.defaultdict(int)
    sessionduration_by_year = collections.defaultdict(float)
    sessionduration_by_day = collections.defaultdict(float)
    visitors_by_year = collections.defaultdict(set)
    visitors_by_day = collections.defaultdict(set)
    for i, (start, end, visitor_id) in enumerate(sessions):
        # print(i, start, end, visitor_id)
        sessions_by_year[start.year] += 1
        sessions_by_day[start.date()] += 1
        sessionduration_by_year[start.year] += (end - start).total_seconds()
        sessionduration_by_day[start.date()] += (end - start).total_seconds()
        visitors_by_year[start.year].add(visitor_id)
        visitors_by_day[start.date()].add(visitor_id)
    print(sessions_by_year)
    print(sessions_by_day)
    print(sessionduration_by_year)
    print(sessionduration_by_day)
    print(visitors_by_year)
    print(visitors_by_day)

    # prepare session count and session duration by year
    labels = list()
    session_count = list()
    session_duration = list()
    for i in range(min(sessions_by_year), max(sessions_by_year) + 1):
        labels.append(i)
        session_count.append(sessions_by_year[i])
        session_duration.append(sessionduration_by_year[i] / sessions_by_year[i] if sessions_by_year[i] > 0 else 0)

    df_sessions_by_year = pandas.DataFrame(zip(labels, session_count), columns=['year', 'count'])
    df_sessionduration_by_year = pandas.DataFrame(zip(labels, session_duration), columns=['year', 'average duration'])
    print(df_sessions_by_year)
    print(df_sessionduration_by_year)

    ax = seaborn.barplot(x="year", y="count", data=df_sessions_by_year)
    matplotlib.pyplot.style.use('dark_background')
    f, axs = matplotlib.pyplot.subplots(1, 2, figsize=(8, 4), dpi=300)
    f.suptitle('Session')
    seaborn.barplot(x="year", y="count", data=df_sessions_by_year, ax=axs[0])
    seaborn.barplot(x="year", y="average duration", data=df_sessionduration_by_year, ax=axs[1])
    axs[1].set_ylabel('duration [s]')
    f.tight_layout()
    f.savefig(os.path.join('..', '..', 'data', 'session_metrics_by_year.svg'))

    # prepare session count and session duration by day
    year = max(sessions_by_year)
    labels = list()
    feature_labels = collections.defaultdict(str)
    feature_labels['session_count'] = 'count'
    feature_labels['session_duration'] = 'duration [s]'
    features = collections.defaultdict(list)
    for date in dates_by_year(year):
        labels.append(date)
        features['session_count'].append(sessions_by_day[date])
        features['session_duration'].append(sessionduration_by_day[date] / sessions_by_day[date] if sessions_by_day[date] > 0 else 0)
    
    for feature in features:
        df = pandas.DataFrame(zip(labels, features[feature]), columns=['date', 'feature'])
        f.clear()
        f, ax = matplotlib.pyplot.subplots(1, 1, figsize=(8, 4), dpi=300)
        f.suptitle('Session')
        seaborn.barplot(x='date', y='feature', data=df, ax=ax)
        ax.set_ylabel(feature_labels[feature])
        x_labels = df['date'][::14].tolist()
        matplotlib.pyplot.xticks(range(0, len(df.index), 14), x_labels)
        f.autofmt_xdate(rotation=90)
        f.tight_layout()
        f.savefig(os.path.join('..', '..', 'data', f'{feature}_by_day.svg'))
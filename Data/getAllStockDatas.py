#coding:utf-8

import csv
import datetime
import time
import sys
import logging
import json
import socket
import os 
import baostock as bs
import pandas as pd
import datetime


if __name__ == '__main__':
    date = datetime.datetime.now()
    datestr = '%d-%02d-%02d' % (date.year, date.month, date.day)
    csv_reader = csv.reader(open('stockList_index.csv', encoding='utf-8'))
    lg = bs.login()
    print('login respond error_code:'+lg.error_code)
    print('login respond  error_msg:'+lg.error_msg)

    for row in csv_reader:
            print(row[0])
            rs = bs.query_history_k_data(row[0].split(".")[1]+ "." + row[0].split(".")[0],"date,open,peTTM,pbMRQ,tradestatus,close,high,low",start_date='2014-01-01', end_date='2021-02-05',frequency="d", adjustflag="3")
            data_list = []
            while (rs.error_code == '0') & rs.next():
                data_list.append(rs.get_row_data())
            result = pd.DataFrame(data_list, columns=rs.fields)
            result.to_csv(row[0].split(".")[0] + "_ALL.csv", index=False)
            print(row[0])

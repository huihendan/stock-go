#coding:utf-8

import time
import os
import datetime
from fnmatch import fnmatch
import shutil
import baostock as bs

import csv
import sys
import logging
import json
import socket
import baostock as bs
import pandas as pd

CAPT_PATH = '/Users/beven/Data/DataEchart/'
Windows = False
if Windows:
    CAPT_PATH = 'D:\\Data_T/'
if not os.path.exists(CAPT_PATH):
    os.mkdir(CAPT_PATH)

def get_last_line(filename):

    try:
        filesize = os.path.getsize(filename)
        if filesize == 0:
            return None
        else:
            with open(filename, 'rb') as fp: # to use seek from end, must use mode 'rb'
                offset = -8                 # initialize offset
                while -offset < filesize:   # offset cannot exceed file size
                    fp.seek(offset, 2)      # read # offset chars from eof(represent by number '2')
                    lines = fp.readlines()  # read from fp to eof
                    if len(lines) >= 2:     # if contains at least 2 lines
                        return lines[-1]    # then last line is totally included
                    else:
                        offset *= 2         # enlarge offset
                fp.seek(0)
                lines = fp.readlines()
                return lines[-1]
    except FileNotFoundError:
        print(filename + ' not found!')
        return None




def test():
    index = 0
    date = datetime.datetime.now()
    datestr = '%d-%02d-%02d' % (date.year, date.month, date.day)
    lg = bs.login()
    print('login respond error_code:'+lg.error_code)
    print('login respond  error_msg:'+lg.error_msg)
    for file in os.listdir(CAPT_PATH):
        if fnmatch(file, 'sh*.csv') or fnmatch(file, 'sz*.csv'):
            lastline = get_last_line(CAPT_PATH + file)
            
            tmp = str(lastline[0:10],'utf-8')
            print(tmp)
            if tmp == 'date,time,' or len(tmp)==0:
               tmp = '2017-04-20'
            lastDay = datetime.datetime.strptime(tmp, "%Y-%m-%d")
            tomorrow = lastDay + datetime.timedelta(days=1)
            startDay = tomorrow.strftime("%Y-%m-%d")
            stock = file.split("_")[0]
            rs = bs.query_history_k_data(stock,"date,open,peTTM,pbMRQ,tradestatus,close,high,low",start_date = startDay, end_date=datestr,frequency="d", adjustflag="3")
            data_list = []
            while (rs.error_code == '0') & rs.next():
                    data_list.append(rs.get_row_data())
            result = pd.DataFrame(data_list, columns=rs.fields)
            result.to_csv(CAPT_PATH + file, mode='a', header=False, index=False)
            print(file)


if __name__ == '__main__':
    test()

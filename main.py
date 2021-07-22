# A script that generates ahr999 index data into a JSON file. 

from pycoingecko import CoinGeckoAPI
from scipy import stats
import math,time,json,datetime

now = round(time.time())
nybf = now - 24*60*60*365*10    #nine year before
cg = CoinGeckoAPI()
scdata = cg.get_coin_market_chart_range_by_id('bitcoin','usd',str(nybf),str(now))    # source data
scdict = {}
for d in scdata['prices']:
    datadate = datetime.datetime.utcfromtimestamp(d[0]/1000).strftime('%Y-%m-%d')
    scdict[datadate]=d[1]                             #存两百天价格和日期的dict

avgut = scdata['prices'][199][0]/1000                             #200天 start unix time
avguts= avgut                                                 #用来被减掉200天，计算平均数
avghm = datetime.datetime.utcfromtimestamp(avgut).strftime('%Y-%m-%d') #200天 human time
avghms = avghm
thps = scdict[avghm]
pcls = []
for i in range(200):
    if avghm in scdict:
        pcls.append(scdict[avghm])
    avgut -= 24*60*60
    avghm = datetime.datetime.utcfromtimestamp(avgut).strftime('%Y-%m-%d') 
 
thavg = round(stats.gmean(pcls),2)       #200天平均值
day = ((avguts- 1230940800) / (24 * 60 * 60) )
logprice = round(10 ** (5.84 * math.log(day, 10) - 17.01),2)
ahr999 = round((thps/thavg)*(thps/logprice),4)
avguts += 24*60*60
avgut = avguts
pcls = []
print("日期:",avghms)
print("200日平均值:",thavg)
print('拟合价格:',logprice)
print('天数:',day)
print("当天价格:",thps)
print("ahr999:",ahr999)

alltimedict={}
while avguts <= now:
    avghm = datetime.datetime.utcfromtimestamp(avgut).strftime('%Y-%m-%d') #200天 human time
    avghms = avghm
    if avghm in scdict:
        thps = scdict[avghm]
    for i in range(200):
        if avghm in scdict:
            pcls.append(scdict[avghm])
        else:
            pass#print("not exist:",avguts)
        avgut -= 24*60*60
        avghm = datetime.datetime.utcfromtimestamp(avgut).strftime('%Y-%m-%d') #200天 human time
    thavg = round(stats.gmean(pcls),2)
    day = ((avguts- 1230940800) / (24 * 60 * 60) )
    logprice = round(10 ** (5.84 * math.log(day, 10) - 17.01),2)
    ahr999 = round((thps/thavg)*(thps/logprice),4)
    
#    print("日期:",avghms)
#    print("ahr999:",ahr999)
#    print("200日平均值:",thavg)
#    print('拟合价格:',logprice)
#    print("当天价格:",round(thps,2))
#    print("\n")
    dt = datetime.datetime.strptime(avghms, "%Y-%m-%d")
    dt = datetime.datetime.timestamp(dt)
    dt *= 1000
    print(avghms, dt)
    alltimedict[dt]=ahr999
    avguts+=24*60*60
    avgut = avguts
    pcls=[]

ls1 = list(alltimedict)
ls2 = list(alltimedict.values())
ls4 = []                                # A list that contains a lot of lists (e.g [[timestamp,price],[timestamp,price],...])

for i in range(len(alltimedict)):
    ls3 = [ls1[i],ls2[i]]
    ls4.append(ls3)

with open("data.json", "w+") as f:      # Check out the data.json file to see the result
    f.write(str(ls4))

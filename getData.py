
#import requests;
import json;
import asyncio;
import aiohttp;
import traceback;
import re;
from pathlib import Path;
from collections import Counter;


#FUNCTIONS
def readKeywords(path):
    techList = []   
    """Read keywords from a text file where each value is on a seperate line"""
    try:
        with open(path) as input_file:
            #stripping newline characters
            techList = [line.rstrip('\n') for line in input_file]
            techList = [line.upper() for line in techList]


    #if an error occurs, print the error.
    except Exception as e: print("Problem with technologies.txt\n--------------------------\n" + str(e))
    if len(techList) < 1:
        raise Exception("No keywords found")
    return techList

async def parseData(postings):
    """Fetch json, parse the data and prepare it to be turned into json"""

    keywordMentions = []

    for i in postings:
        s = i['descr'].upper()
        heading = i["heading"]
        company = i["company_name"]
        #Remove zipcodes and '-' signs using regex
        location = re.sub(r"\d+|-+", '', str(i["municipality_name"])).strip()
        if(location == ""):
            location = "none";
        link = "https://duunitori.fi/tyopaikat/tyo/" + i["slug"]

        #List of keywords found in post
        regexFound = re.findall(r"\bC\+\+|\bC\#|\b" + r'\b|\b'.join(KWList) + r"\b", s, re.IGNORECASE);
        keywordsFound = list(set(regexFound))

        #keywordsFound = list({x for x in KWList if bool(re.search(r"\b" + x + r"\b", s, re.IGNORECASE))})

        for i in range(len(keywordsFound)):
            #strip whitespace and remove , and ) characters from the matches
            keywordsFound[i] = keywordsFound[i].strip()
            keywordsFound[i] = keywordsFound[i].replace(",", "").replace(")", "").replace("-", "").replace("(", "")

        job = {"heading": heading, "link": link, "technologies": list(set(keywordsFound)), "company": company, "location": location}

        for keyword in keywordsFound:
            if (keyword in technologies):
                technologies[keyword]['jobs'].append(job)
            else:
                technologies[keyword] = dict()
                technologies[keyword]['name'] = keyword
                technologies[keyword]['jobs'] = []
                technologies[keyword]['jobs'].append(job)


        #create company and location dictionaries
        if len(keywordsFound) > 0:
            if company in companies:
                companies[company]["technologies"] += keywordsFound
                companies[company]["jobs"].append(job)

            else:
                companies[company] = dict()
                companies[company]["name"] = company
                companies[company]["technologies"] = keywordsFound
                companies[company]["jobs"] = []
                companies[company]["jobs"].append(job)

            if location in locations:
                locations[location]["technologies"] += keywordsFound
                locations[location]["jobs"].append(job)
            else:
                locations[location] = dict()
                locations[location]["name"] = location
                locations[location]["technologies"] = keywordsFound
                locations[location]["jobs"] = []
                locations[location]["jobs"].append(job)

            keywordMentions += keywordsFound
    return keywordMentions

#Fetch the postings site json
async def fetch(url):
    """Execute an http call async
    Args:
        url: URL to call
    Return:
        responses: A dict like object containing http response
    """
    async with aiohttp.ClientSession(headers={"accept": "application/json", "User-Agent": "Curl/7.64.1"}) as session:
        async with session.get(url) as response:
            resp = await response.json()
            print(resp)
            return resp

#fetch all postings
async def handleData():
    tasks = []
    allMentionsFound = []
    #call api to get all postings
    url = "https://duunitori.fi/api/v1/jobentries?search=koodari&search_also_descr=1&format=json"
    results = await fetch(url) 
    print("results gotten from: " + url)
    count = results["count"]
    posts = results["results"]
    tasks.append(parseData(results['results']))
    nextPage = results['next']

    while nextPage != None:
        url = nextPage
        results = await fetch(url)
        print("results gotten from: " + url)
        posts += results["results"]
        tasks.append(parseData(results['results']))
        nextPage = results['next']


    jsonDataToFile({"count": count, "posts": posts}, Path('./json/posts.json'))

    mentions = (await asyncio.gather(*tasks, return_exceptions=False))
    for m in mentions:
        allMentionsFound += m


    return allMentionsFound, count

def formatAndCreateJson(count, kwMentions):
    """Format and create a usable json file from data"""
    global companies
    global locations
    techCount = dict()
    techCount = Counter(kwMentions)
    for key in techCount:
        technologies[key]['jobs_count'] = techCount[key]

    for key in companies:
        companies[key]["technologies"] = Counter(companies[key]["technologies"])
        companies[key]["jobs_count"] = len(companies[key]["jobs"])


    for key in locations:
        locations[key]["technologies"] = Counter(locations[key]["technologies"])
        locations[key]["jobs_count"] = len(locations[key]["jobs"])
    
    data["posts_count"] = count
    data["technologies"] = technologies
    data["companies"] = companies
    data["locations"] = locations

    jsonDataToFile(data, Path('./json/data.json'))




def jsonDataToFile(d, path):
    """Creates/overwrites data to a file in json format"""
    try:
        jsonData = json.dumps(d)

        f = open(path, "w+")
        f.write(jsonData)
    except Exception as e: print(f"An exception occurred while saving json data to file {path}:\n {e}")


#VARIABLES
KWList = []
technologies = dict()
companies = dict()
locations = dict()
data = dict()

try:
    KWList = readKeywords(Path('./keywords/technologies.txt'))
    allMentions, count = asyncio.run(handleData())
    formatAndCreateJson(count, allMentions)
except Exception:
    print(traceback.format_exc())


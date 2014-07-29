#!/usr/bin/python
#
# Copyright 2012 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import httplib2
from apiclient.discovery import build
from collections import Counter
from oauth2client import tools
from oauth2client.client import OAuth2WebServerFlow
from oauth2client.file import Storage
from oauth2client.tools import run_flow

# For these examples, the client id and client secret are command-line arguments
parser = argparse.ArgumentParser(description=__doc__,
    formatter_class=argparse.RawDescriptionHelpFormatter,
    parents=[tools.argparser])
parser.add_argument('client_id',
                    help='A Google "Client ID for native application" that '
                         'has the Genomics API enabled.')
parser.add_argument('client_secret',
                    help='The client secret that matches the Client ID.')
flags = parser.parse_args()

# Authorization
storage = Storage('credentials.dat')
credentials = storage.get()
if credentials is None or credentials.invalid:
  flow = OAuth2WebServerFlow(flags.client_id, flags.client_secret,
                             'https://www.googleapis.com/auth/genomics')
  credentials = run_flow(flow, storage, flags)

# Create a genomics API service
http = httplib2.Http()
http = credentials.authorize(http)
service = build('genomics', 'v1beta', http=http)


#
# This example gets the read bases for NA12878 at specific a position
#
dataset_id = 376902546192 # This is the 1000 Genomes dataset ID
reference_name = '22'
reference_position = 51005354

# 1. First find the readset ID for NA12878
request = service.readsets().search(
  body={'datasetIds': [dataset_id], 'name': 'NA12878'},
  fields='readsets(id)')
readsets = request.execute().get('readsets', [])
if len(readsets) != 1:
  raise Exception('Searching for NA12878 didn\'t return '
                  'the right number of results')

na12878 = readsets[0]['id']


# 2. Once we have the readset ID,
# lookup the reads at the position we are interested in
request = service.reads().search(
  body={'readsetIds': [na12878],
        'sequenceName': reference_name,
        'sequenceStart': reference_position,
        'sequenceEnd': reference_position,
        'maxResults': '1024'},
  fields='reads(position,originalBases,cigar)')
reads = request.execute().get('reads', [])

# Note: This is simplistic - the cigar should be considered for real code
bases = [read['originalBases'][reference_position - read['position']]
         for read in reads]

print 'NA12878 bases on %s at %d' % (reference_name, reference_position)
for base, count in Counter(bases).items():
  print '%s: %s' % (base, count)

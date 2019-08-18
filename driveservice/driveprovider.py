from __future__ import print_function
import pickle
import os.path
from googleapiclient.discovery import build
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request

class DriveProvider:

    SCOPES = ['https://www.googleapis.com/auth/drive.metadata.readonly']

    def authorize(self):
        creds = None

        if os.path.exists('token.pickle'):
            with open('token.pickle', 'rb') as token:
                creds = pickle.load(token)
        
        if not creds or not creds.valid:
            if creds and creds.expired and creds.refresh_token:
                creds.refresh(Request())
            else:
                flow = InstalledAppFlow.from_client_secrets_file('credentials.json', self.SCOPES)
                creds = flow.run_local_server(port=0)
            
            with open('token.pickle', 'wb') as token:
                pickle.dump(creds, token)
        
        return creds

    def listLastModified(self, creds, pageSize=10):
        service = build('drive', 'v3', credentials=creds)
        results = service.files().list(pageSize=pageSize, orderBy='modifiedTime desc', fields="nextPageToken, files(id,name,parents)").execute()
        items = results.get('files', [])

        for item in items:
            parent = service.files().get(fileId=item['parents'][0]).execute()
            print("{0}\\{1}".format(parent['name'], item['name']))


provider = DriveProvider()
credentials = provider.authorize()
provider.listLastModified(credentials)
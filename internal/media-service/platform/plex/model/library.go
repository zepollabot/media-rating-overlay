package plex

type Library struct {
	ID       string `json:"key"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Language string `json:"language"`
	Hidden   int    `json:"hidden"`
}

/*
   "Directory": [
       {
           "allowSync": false,
           "art": "/:/resources/movie-fanart.jpg",
           "composite": "/library/sections/6/composite/1678577545",
           "filters": true,
           "refreshing": false,
           "thumb": "/:/resources/movie.png",
           "key": "6",
           "type": "movie",
           "title": "Anime Movies",
           "agent": "tv.plex.agents.movie",
           "scanner": "Plex Movie",
           "language": "it-IT",
           "uuid": "9908faef-610f-4c43-a2ce-b11f43c9aa66",
           "updatedAt": 1674841929,
           "createdAt": 1633087805,
           "scannedAt": 1678577545,
           "content": true,
           "directory": true,
           "contentChangedAt": 1425204,
           "hidden": 0,
           "Location": [
               {
                   "id": 6,
                   "path": "/share/CACHEDEV1_DATA/Multimedia/Anime OAV"
               }
           ]
       },
*/

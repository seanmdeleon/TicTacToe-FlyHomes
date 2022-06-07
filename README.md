# TicTacToe-FlyHomes
TicTacToe coding challenge for FlyHomes interview

--> To Start the localhost HTTP server <--

    In a terminal window, naviate to the /main folder and run './main' for the compiled executable. The HTTP server will turn on and remain idle until closed
    via CTRL+C
    Keep this terminal open and the main program running while utilizing the API for game play
    If you wish to compile on your own, run 'go build' in this folder to recompile the main executable

--> To Play a game <--

    While the localhost http server is running in one terminal window, open a second terminal window to send HTTP requests using cURL.
    Using the same machine, people can play multiple games using multiple terminal windows as the API can handle multiple requests in parallel

    The following is the list of cURL commands one can use for each of the endponts

    GET tictactoe/
        Return all games InProgress
    
        curl -v http://localhost:8080/tictactoe

        Example Response
		    {"games": ["5fb190f-20d7-4a3f-beef-6191342ae06a", "e8d50f36-25fb-49ff-85d2-aa516cf6327b"] }

    POST tictactoe/
        Create a new game

        curl -v --header "Content-Type: application/json" -d "{\"players\":[\"player1\", \"player2\"], \"columns\": 3, \"rows\": 3}" http://localhost:8080/tictactoe

        Example Response
            {"gameId":"e8d50f36-25fb-49ff-85d2-aa516cf6327b"}
    
    GET tictactoe/{game_id}
        Get a game with game_id

        curl -v http://localhost:8080/tictactoe/c2b9352d-ded2-4177-a38a-d54df68d32d3

        Example Response
            {"players":["player1","player2"],"state":"IN_PROGRESS"}
    
    GET tictactoe/{game_id}/moves
        Get a list or sublist of moves for a give game_id
        start and until are optional

        curl -v http://localhost:8080/tictactoe/e5fb190f-20d7-4a3f-beef-6191342ae06a/moves?start=0&until=1
    
    GET tictactoe/{game_id}/moves/{move_number}
        Get a move for a game_id with a move_number, move_number is 0 offset

        curl -v http://localhost:8080/tictactoe/e5fb190f-20d7-4a3f-beef-6191342ae06a/moves/2

        Example Response
            {"type":"MOVE","player":"player2","row":0,"col":1}
    
    POST tictactoe/{game_id}/{player_id}
        Post a Move
        playerID is either 0 or 1, unique per game_id

        curl -v --header "Content-Type: application/json" -d "{\"row\": 1, \"column\": 1}" http://localhost:8080/tictactoe/e5fb190f-20d7-4a3f-beef-6191342ae06a/0

        Example Response
            {"move":"c2b9352d-ded2-4177-a38a-d54df68d32d3/moves/4"}
    
    PUT tictactoe/{game_id}/quit
        Quit a game provided the game_id

        curl -v -X PUT http://localhost:8080/tictactoe/c2b9352d-ded2-4177-a38a-d54df68d32d3/quit

        Example Response
            {"quitGame":"e8d50f36-25fb-49ff-85d2-aa516cf6327b"}

--> Design Thoughts by Sean <--

    This project took me longer than expected, but I still enjoyed it! Because work is busy, I made some decisions to make my submission simple. I could have easily made this project super airtight and user friendly, but I didn't have enough time in my day. I would like to talk about a more sophisticated, well maintained approach in the followup interview

    1) Database. Truly InMemory, where the DB lives within the scope of the http server's memory space. I recognize that storing the backend data within a Go map
        quickly communicates the idea of a an InMemory DB, but the program must remain on in order to keep all data alive. It would have been even more interesting to create a second application that controls the InMemory DB running on a second server. This second server would run on localhost using a port other than 8080. I found that for the sake of ease and brevity, storing the InMemory DB within the same application as the REST api was the way to go. I implemented atomic read/write for the InMemory DB table to maintain some level of integrity
    
    2) Validation. I tried to validate a fair amount of edge cases for each endpoint using a popular validation package within Golang. Input validation could be endless if desired, I
        stopped adding further validation/error details once I felt like a fair basis was covered.
    
    3) Golang. Probably not the best language to code a REST API. It takes a long time, and I feel like python/Django is more sophisticated. Golang, however, is the language I know
        best, and I have written an API in Golang before at my current job
    
    4) Testing. I wanted to add some tests, but not many, given how much time I was able to budget on this coding assignment (Work has been busy!). Normally my unit tests would be much more exhaustive. I gave a couple of tests just to showcase that I think about testing and mocking

    5) UI. I spoke with Joel and he said there was no need to also build a UI. I didn't have time to create a barebones GUI, only enough time to create the REST API. I tried to write this backend code with the idea that a UI dev could easily interract with the service programtically, even including verbose comments that describe expected behavior of endpoints

    6) Scalability.  I understand my implementation as is does not scale, as it is a local appplication. I really wanted to build this whole project using AWS resources in my personal AWS account as a way to showcase my ability to use AWS. API Gateway to route HTTP requests, AWS Lambda to serve the requests routed from API Gateway, Dynamodb to store the data. To scale the application even futher to serve many users, I would love to discuss my thoughts in the follow up interview (in addition to the AWS resources mentioned, utilize caching, shard data, etc). I would utiltize terraform as a way to keep track of all resources

# [e-inwork.com](https://e-inwork.com)

## Getting Started - Golang Team Microservice
### Run the Golang Team Microservice and the [the Golang User Microservice](https://github.com/e-inwork-com/go-user-service)
#### The application will be running with two different microservices in one port 8000 using [Envoy](https://www.envoyproxy.io).
1. Install Docker
    - https://docs.docker.com/get-docker/
2. Git clone this repository to your localhost, and from the terminal run below command:
   ```
   git clone git@github.com:e-inwork-com/go-team-service
   ```
3. Change the active folder to `go-team-service`:
   ```
   cd go-team-service
   ```
4. Run Docker Compose:
   ```
   docker-compose up -d
   ```
5. Create a user in the User API with CURL command line:
    ```
    curl -d '{"email": "jon@doe.com", "password": "pa55word", "first_name": "Jon", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users
    ```
6. Login to the User API:
   ```
   curl -d '{"email":"jon@doe.com", "password":"pa55word"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users/authentication
   ```
7. You will get a `token` from the response login, and set it as a token variable for example like below:
   ```
   token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjhhY2NkNTUzLWIwZTgtNDYxNC1iOTY0LTA5MTYyODhkMmExOCIsImV4cCI6MTY3MjUyMTQ1M30.S-G5gGvetOrdQTLOw46SmEv-odQZ5cqqA1KtQm0XaL4
   ```
8. Create a team for current user, you can use any image or use the image on the folder test:
   ```
   curl -F team_name="Doe's Team" -F team_picture=@/YourRootFolder/.../go-team-service/api/test/team.jpg -H "Authorization: Bearer $token"  -X POST http://localhost:8000/service/teams
   ```
9. The response will show a team picture, open it on the browser for example like this http://localhost:8000/service/teams/pictures/926d610c-fd54-450e-aa83-030683227072.jpg
10. Get the team of the current user:
  ```
  curl -H "Authorization: Bearer $token" http://localhost:8000/service/teams/me
  ```
11. Set a `team_member_team` variable from the id of the current team for example like below:
   ```
   team_member_team=9a584a86-1212-41f9-8d32-9911ee3f4550
   ```
12. Regsiter a new user for adding it as team member:
  ```
  curl -d '{"email": "nina@doe.com", "password": "pa55word", "first_name": "Nina", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users
  ```
11. Set a `team_member_user` variable from the id of the last user (No. 12) for example like below:
   ```
   team_member_user=08a6c56a-d3f9-47d3-81ea-ced54153e1df
   ```
13. Create a team member:
  ```
  curl -d '{"team_member_team": "'$team_member_team'", "team_member_user": "'$team_member_user'"}' -H "Authorization: Bearer $token" -H "Content-Type: application/json" -X POST http://localhost:8000/service/teams/members
  ```
14. Set a `team_member_id` variable from the current team member, for example:
  ```
  team_member_id=35fc84e9-96dd-422b-adf5-e19bba1ec8a4
  ```
15. Get a list team member of teh current user:
  ```
  curl -H "Authorization: Bearer $token" http://localhost:8000/service/teams/members
  ```
16. Delete the current team member, and the response will be `HTTP/1.1 200 OK`:
  ```
  curl -I -H "Authorization: Bearer $token"  -X DELETE http://localhost:8000/service/teams/members/$team_member_id
  ```
17. Have fun!
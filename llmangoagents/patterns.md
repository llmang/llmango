# Patterns for our agent setup

EA - Entry Agent
- There is always a first agent to recieve this message
 -- Often: Organizer, Router, Cheaper
 --in simple flows its just the default agent

Exit Agents(s)
    - At the end of a task we want exit agents who:
        - Grade the work
        - Format the results into a structured formatter
        - Summarize past results
        - route to next path, input to db, etc etc. 

Typical flow for task. 
1. entry agent Recieves Task Logic and decides on what actions to takes
2. Calls tools and Agents 
3. Recieves back all data and either passses it on to the next agent or calls more tools
4. if it "accepts" the result it moves to the exit agents
5. exit agent may 1st grade the work and return it if it finds issues "missing y" "doesnt have z" "incorrect b" and return it to organizer or send it to correction agent. 
6. Once correction agent returns it again grades it.
7. It proceeds to the formating agent for processing. 
8. It proceeds to the next step in flow, input in DB, or returned to user. 



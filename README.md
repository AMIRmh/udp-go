# Reliable UDP 

this project is written in GO

## The idea

the idea is very easy! send the packet and wait for the ack:) but obviously with threads!!!

## The challenges

As UDP does not have any response, it means that the packets are just sent and the sender is not waiting for the response.
so this should be handled by myself, I decided to create a list to wait for the response and those packets that their response were back, would be
removed from the list.



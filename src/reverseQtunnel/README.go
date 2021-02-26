	# reverse tunnel

	With tunnel, client get user request and connect to server, then things happen; 
	With reverse tunnel, user request server, and server should 'connect' client to make things happen;
	Be in local network , client never get connected. 
	To make things happen , we keep a addtional command connection , when user request server, server use this connection tell client to connect server, server accept this connection , then tunnel build up.

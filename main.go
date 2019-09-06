// townstats returns information about tilde.town in the tilde data protcol format
// It was originally a Python script written by Michael F. Lamb <https://datagrok.org>
// License: GPLv3+

// TDP is defined at http://protocol.club/~datagrok/beta-wiki/tdp.html
// It is a JSON structure of the form:

// Usage: stats > /var/www/html/tilde.json

// {
//   'name':         (string) the name of the server.
//   'url':          (string) the URL of the server.
//   'signup_url':   (string) the URL of a page describing the process required to request an account on the server.
//   'want_users':   (boolean) whether the server is currently accepting new user requests.
//   'admin_email':  (string) the email address of the primary server administrator.
//   'description':  (string) a free-form description for the server.
//   'users': [      (array) an array of users on the server, sorted by last activity time
//   	{
//   		'username': (string) the username of the user.
//   		'title':    (string) the HTML title of the user’s index.html page.
//   		'mtime':    (number) a timestamp representing the last time the user’s index.html was modified.
//   		},
//   	...
//   	]
//   'user_count':   (number) the number of users currently registered on the server.
// }

// We add some town-flavored info as well:

// {
//    'users': [ (array) of users on the server.
//    {
//    	'default': (boolean) Is the user still using their unmodified default index.html?
//    	'favicon': (string) a url to an image representing the user
//    	},
//    ...
//    ]
//    'live_user_count': (number) count of live users (those who have changed their index.html)
//    'active_user_count': (number) count of currently logged in users
//    'generated_at': (string) the time this JSON was generated in '%Y-%m-%d %H:%M:%S' format.
//    'generated_at_msec': (number) the time this JSON was generated, in milliseconds since the epoch.
//    'uptime': (string) output of `uptime -p`
//    'news': collection of tilde.town news entries containing 'title', 'pubdate', and 'content', the latter being raw HTML
//  }
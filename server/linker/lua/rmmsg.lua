
local cid = KEYS[1]
local msgid = ARGV[1]

local mk = 'M.'..string.sub(cid, 1, 5)..'.'..msgid
--local mk = 'M.'..cid..'.'..msgid
local smk = 'SM.'..cid

redis.call('DEL', mk)


redis.call('ZREM', smk, msgid)


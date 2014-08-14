

local cid = KEYS[1]
local ck = 'I.'..cid
local remote = ARGV[1]
local restaddr = ARGV[2]
local stamp = ARGV[3]

local vs = redis.call('HMGET', ck,
                      'remote', 
                      'restaddr'
)

if vs[1] == remote and vs[2] == restaddr then
   redis.call('HMSET', ck,
              'remote', "",
              'restaddr', "",
              'closestamp', stamp
   )

end





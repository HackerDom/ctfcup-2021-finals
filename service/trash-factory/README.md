
KEY (8 bytes)
ENCR_DATA (\x03\x13\x37 + MSG_TYPE (1 byte) + data)

1. Create container
   \x01 - operation type
   ARGS:
        \x00 - container size (item counts)
        \x00 x50 - description
2. Container list
   \x02
3. Get container info
   \x03
   \x01 -- container id
4. put item
   \x04
   \x31 -- type
   \x00\x00\x00\x01 -- weight (kg)
   \x00 x50 -- description
5. get item
   \x05
   \x00 -- container id
   \x01 -- item n in contained


Status code:
\x01 - incorrect signature
\x02 - function not found

-----

0. Key == encr token

1. Rand GO same
   rand.Seed(time.Now().Unix())
   fmt.Println(rand.Int())
   
2. Get container info
    container_id <- ../users/<key>.dat
    item_id <- 2

-----

db
|-> users
|     |-> <key>.dat
|-> containers
      |-> <container_id>.dat


<key>.dat
1. ENCR_TOKEN
2. description
3. container_counts
4. container_id_1
5. container_id_2 
   ... 
n. container_id_n-3

<container_id>.dat
1. container_name
2. item_1
3. item_2
   ...
n. item_n-1

WHERE item_n:
    |-> type (1 byte):
        \x31 - plastic
        \x32 - paper
        \x33 - ...
    |-> wight (4 bytes)
    |-> description (50bytes)

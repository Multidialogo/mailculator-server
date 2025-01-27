Pay attention standard email files best practice is to use MicroSoft Windows style carriage returns "\r\n",most *nix 
like systems uses "\n".

DO NOT ALTER CARRIAGE RETURN IN EXPECTATIONS IN THIS DIRECTORY!

Hint:
sed -i 's/$/\r/' expectation.filename.EML should fix any problem
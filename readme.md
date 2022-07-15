# Duolingo Dictionary Reconstructor
 
This tool allows to partially bring back the discontinued Duolingo dictionary. It utilizes the fact that every sentence with user comments had its own automatically created forum post. The program first scans the list of forum posts associated with a given language and collects the IDs of posts which are related to sentences. Because the forum was discontinued as well, this tool sources the data from [an archive](https://duolingo.hobune.stream/) of it. For each post, the utility has to fetch its data in order to get the translation of the sentence. It's done by making requests to the official Duolingo forum because it still contains sentence comments and I want to reduce the network traffic to the archive. Sometimes the requests are unsuccessful, even though when visiting the same URL using a web browser returns the right data. In such cases, the addon falls back to using the archive. Still, there may be hundreds of sentences for which it was impossible to find a translation (so they are either only in the native or only in the target language).

## Usage
First, pick an ID of the forum topic (category) dedicated to the language that you are interested in, using the file `topic_ids.txt`. Then use the compiled executable like this (Linux example):
```shell
# In topic_ids.txt: [en->fr] French -> ID: 147
./dictrec 147 fr > french.txt
```
This will save French sentences in the file named `french.txt`. It should take about half a minute. The visible output will look like this:
```
Downloading...
  Error ("My husband is a good man.", 1911635): <nil> 404 Not Found
Remaining: 1736 / 34248
  Error ("I cannot open the file.", 32233580): <nil> 404 Not Found
  Error ("It is not sand.", 1427726): <nil> 404 Not Found
[...]
Finished: 3065 translations missing
```
The errors mean that the translations of these sentences couldn't be retrieved. Either they were not archived and not available on the official forum or they never had any comments.

The output saved to the file has this format:
```
sentence = translation
sentence without translation = ?
```

Now you can search the file for the words or sentences that you would like to see. A shortcoming, compared to the original dictionary, is that it is significantly harder to search for all forms of a given word at once.

Bear in mind that the sentences are copyrighted, and you should only use them for your own needs.
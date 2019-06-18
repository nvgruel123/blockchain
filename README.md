# Documentation

## Description

  ### This project is a seafood supply chain management using blockchain technology. 

  The reason for using blockchain is it can provide traceability to the supplies sold in the market. Especially for seafood, people usually cares about how fresh the fish they bought. By using blockchain, it gives buyers a direct way of track the seafood in the market. 

## Functionalities / Success

  1. Fishermen can create blocks contain the fish they caught with details such as type, count, weight, date being caught and so on. 
  2. Buyers can view all these informations to make sure they got the fresh seafood they expected. 

## Design

### Disgn choice 
  1. Genesis Block
  - All miners' public keys will be pre-stored in the genesis block. In this way, we can have some control over the miners in the chain. 
    - Fishermen can ask for the program owner to join it by given private key. 
    - There are maximum number of miners in the blockchain according to public keys. 
    - If the number of miners exceed the maximum number, it is possible to start another seperate chain. 

  2. Canonical Chain
    - There will be no canonical chain in this project. 
    - Ideally, miners will produce the next block after the latest block in the chain. But if somehow there are blocks with the same height being produced, all of them will be considered as valid blocks. 

  3. Block Creation and Validation
    - Fishermen can create block to record their caught at any given time. They will also put the height and hash of the block they created on the fish label. 
    - the date fishermen caught the fish should be less than the block created time. 
    - When other miners receive a block, time difference between the block created time and their current time should be less than 30 min. Otherwise, it is not a valid block and will not be accepted by majority of the miners. 

  4. View Block
    - Buyers can go to certain website that hosting by the supply chain owner and insert block height and hash they got from fish's label. The website will display the information related to the seafood.

### Implementation

  1. This project will use blockchain server from project 4 as backend API.

  2. It will use react to provide a UI for miners and buyers. 

## Video Demo Link
  [demo](https://drive.google.com/file/d/1pgZvE5f3S7gOB9NDsqQr24lcx2CI14nb/view?usp=sharing)
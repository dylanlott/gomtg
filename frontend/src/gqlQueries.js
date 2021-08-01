import gql from 'graphql-tag';

export const boardstates = gql`
query($gameID: String!) {
  boardstates(gameID: $gameID) {
    User {
      Username
      ID
    }
    Life
    GameID
    Commander {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Library {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Graveyard {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Exiled {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID   
    }
    Field {
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Hand {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Revealed {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Controlled {
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
  }
}
`

export const updateBoardStateQuery = gql`
  mutation ($boardstate: InputBoardState!) {
    updateBoardState(input: $boardstate) {
      User {
        Username
        ID
      }
      GameID
      Life
      Commander { 
        Name 
        ID 
        Tapped
        Flipped
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Library { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Graveyard { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Exiled { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Field { 
        Name 
        ID 
        Tapped
        Flipped
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Hand { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Revealed { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Controlled { 
        Name 
        ID 
        Tapped
        Flipped
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      } 
    }
  }
`

export const boardstatesSubscription = gql`
subscription($boardstate: InputBoardState!) {
  boardUpdate(boardstate: $boardstate) {
    User {
      Username
    }
    Life
    Commander {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Library {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Graveyard {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Exiled {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID   
    }
    Field {
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID  
    }
    Hand {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Revealed {
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Controlled {
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
  }
}
`

export const selfStateQuery = gql`
  query($gameID: String!, $userID: String) {
    boardstates(gameID: $gameID, userID: $userID) {
      User {
        Username
      }
      Life
      Commander { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Library { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Graveyard { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Exiled { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Field { 
        Name 
        ID 
        Tapped
        Flipped
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Hand { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Revealed { 
        Name 
        ID 
        Tapped
        Flipped
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      }
      Controlled { 
        Name 
        ID 
        Colors 
        ColorIdentity 
        ManaCost 
        Power 
        Toughness 
        CMC 
        Text 
        Types 
        Subtypes 
        Supertypes 
        IsTextless 
        TCGID 
        ScryfallID 
      } 
    }
  }
`

// gameUpdateQuery powers the TurnTracker and Opponents components.
export const gameUpdateQuery = gql`
subscription($gameID: String!, $userID: String!) {
  gameUpdated(gameID: $gameID, userID: $userID) {
    ID
    PlayerIDs {
      Username
      ID
    }
    Turn {
      Player
      Phase
      Number
    }
  }
} 
`

export const boardstateSubscription = gql`
subscription($obsID: String!, $userID: String!) {
  boardstateUpdated(observerID: $obsID, userID: $userID) {
    User {
      ID
      Username
    }
    Life
    GameID
    Commander { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Library { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Graveyard { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Exiled { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Field { 
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Hand { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Revealed { 
      Name 
      ID 
      Tapped
      Flipped
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    }
    Controlled { 
      Name 
      ID 
      Colors 
      ColorIdentity 
      ManaCost 
      Power 
      Toughness 
      CMC 
      Text 
      Types 
      Subtypes 
      Supertypes 
      IsTextless 
      TCGID 
      ScryfallID 
    } 
  }
}
`

export const gameQuery = gql`
query ($gameID: String) {
  games(gameID: $gameID) {
    ID
    PlayerIDs {
      Username
      ID
    }
    Turn {
      Player
      Phase
      Number
    }
  }
}
`

export const updateGame = gql`
mutation ($input: InputGame!) {
  updateGame(input: $input) {
    ID
    PlayerIDs {
      ID
      Username
    }
    Turn {
      Phase
      Player
      Number
    }
  }
}
`

export const signup = gql`
mutation ($username: String!, $password: String!) {
  signup(username: $username, password: $password) {
    ID
    Username
    Token
  }
}
`

export const login = gql`mutation($username: String!, $password: String!) {
  login(username: $username, password: $password) {
    Username
    ID
    Token 
  }
}`
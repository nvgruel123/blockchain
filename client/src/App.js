import React from 'react';
import { BrowserRouter, Route, Switch} from "react-router-dom";
import Main from './Main';

function App() {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/" component = {Main}>

        </Route>
      </Switch>
    </BrowserRouter>
  );
}

export default App;

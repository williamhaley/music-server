import React from 'react';
import './App.css';
import Albums from './Albums';
import Store from './Store';
import Player from './Player';
import { HashRouter, Switch, Route, Redirect, RouteComponentProps } from 'react-router-dom';
import AlbumDetails from './AlbumDetails';
import { AuthButton, LoginPage, PrivateRoute, ProvideAuth } from './useAuth';

function App() {
  return (
    <ProvideAuth>
      <HashRouter>
        <Store>
          <>
            <AuthButton />

            <Switch>
              <Route path="/login">
                <LoginPage />
              </Route>

              <PrivateRoute path="/albums/:id" render={({ match }: RouteComponentProps) => {
                const params = match.params as { [id: string]: string };
                const id = params.id as string;
                return <AlbumDetails id={id} />;
              }} />

              <PrivateRoute path="/albums">
                <Albums />
              </PrivateRoute>

              <PrivateRoute path="/">
                <Redirect to="/albums" />
              </PrivateRoute>
            </Switch>

            <Player />
          </>
        </Store>
      </HashRouter>
    </ProvideAuth>
  );
}

export default App;

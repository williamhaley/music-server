import React, { useContext, createContext, useState, useEffect, FormEvent } from "react";
import {
  Route,
  Redirect,
  useHistory,
  useLocation
} from "react-router-dom";

const fakeAuth = {
  isAuthenticated: false,

  async signin(token: string): Promise<void> {
    const clientId = Array.from(crypto.getRandomValues(new Uint8Array(32)), dec => dec.toString(16).padStart(2, "0")).join('');
    localStorage.setItem('clientId', clientId);

    const res = await fetch(`/api/auth?Access-Token=${token}`);
    if (res.status !== 200) {
      throw new Error('invalid token');
    }

    localStorage.setItem('token', token);
    fakeAuth.isAuthenticated = true;
  },

  async signout() {
    localStorage.setItem('token', '');
    fakeAuth.isAuthenticated = false;
  }
};

type AuthState = {
  loading: boolean;
  token: string | null;
}

type AuthContext = {
  authState: AuthState;
  signin: (token: string) => Promise<void>;
  signout: () => Promise<void>;
}

const authContext = createContext<AuthContext>({
  authState: { token: null, loading: true },
  signin: () => Promise.resolve(),
  signout: () => Promise.resolve(),
});

type ProvideAuthProps = {
  children: JSX.Element;
}

const useProvideAuth = (): AuthContext => {
  const [authState, setState] = useState<AuthState>({ loading: true, token: null });

  const signin = async (token: string) => {
    setState({ token: null, loading: true });
    await fakeAuth.signin(token);
    setState({ token, loading: false });
  };

  const signout = async () => {
    await fakeAuth.signout();
    setState({ token: null, loading: false });
  };

  useEffect(() => {
    if (localStorage.getItem('token')) {
      signin(`${localStorage.getItem('token')}`);
    } else {
      signout();
    }
  }, []);

  return {
    authState,
    signin,
    signout
  };
}

function ProvideAuth({ children }: ProvideAuthProps) {
  const auth = useProvideAuth();

  return (
    <authContext.Provider value={auth} >
      {children}
    </authContext.Provider>
  );
}

function useAuth() {
  return useContext(authContext);
}

function AuthButton() {
  let history = useHistory();
  let auth = useAuth();

  if (auth.authState.loading) {
    return <p>Loading...</p>;
  }

  return auth.authState.token ? (
    <p>
      Welcome!{" "}
      <button
        onClick={async () => {
          await auth.signout();
          history.push("/");
        }}
      >
        Sign out
      </button>
    </p>
  ) : (
    <p>You are not logged in.</p>
  );
}

const PrivateRoute = ({ ...rest }) => {
  let auth = useAuth();

  if (!auth.authState.token) {
    return (
      <Route render={({ location }) => (
        <Redirect
          to={{
            pathname: "/login",
            state: { from: location }
          }}
        />
      )} />
    );
  }

  return (
    <Route {...rest} />
  );
}

interface LocationState {
  from: {
    pathname: string;
  };
}

const LoginPage = () => {
  let history = useHistory();
  let location = useLocation<LocationState>();
  let auth = useAuth();

  if (auth.authState.token) {
    history.push("/");
    return null;
  }

  let { from } = location.state || { from: { pathname: "/" } };

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!event.target) {
      return;
    }

    const formData = new FormData(event.target as HTMLFormElement);
    const token = `${formData.get('token')}`;

    try {
      await auth.signin(token);
      history.replace(from);
    } catch (error) {
      alert('invalid code or error occurred');
    }
  };

  if (auth.authState.loading) {
    return <p>...</p>;
  }

  return (
    <form onSubmit={onSubmit}>
      <p>Enter a valid access token</p>
      <input type="text" name="token" required />
      <button type="submit">Log in</button>
    </form>
  );
}

export { ProvideAuth, AuthButton, PrivateRoute, LoginPage, useAuth };
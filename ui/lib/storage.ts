const TOKEN_KEY = "wego_token";

export function storeToken(token: string) {
  if (window && window.localStorage) {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function getToken() {
  if (window && window.localStorage) {
    return localStorage.getItem(TOKEN_KEY);
  }

  return null;
}

export function removeToken() {
  if (window && window.localStorage) {
    localStorage.removeItem(TOKEN_KEY);
  }
}

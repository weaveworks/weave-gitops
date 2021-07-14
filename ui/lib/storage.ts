const TOKEN_KEY = "wego_token";

export function storeToken(token: string) {
  if (window && window.sessionStorage) {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function getToken() {
  if (window && window.sessionStorage) {
    return localStorage.getItem(TOKEN_KEY);
  }

  return null;
}

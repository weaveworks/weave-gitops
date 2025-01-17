// TokenRefreshWrapper is a singleton service wrapper calls a user-provided
// token refresh function if a 401 is returned
//
// It only does a single refresh at a time, requests that are made during
// the refresh wait for the refresh to finish first.
export class TokenRefreshWrapper {
  private static refreshPromise: Promise<void> | null = null;
  private static refreshTokenFn: () => Promise<void>;

  static wrap(api: any, refreshTokenFn: () => Promise<void>) {
    this.refreshTokenFn = refreshTokenFn;

    const wrapped: any = {};
    for (const method of Object.getOwnPropertyNames(api)) {
      if (typeof api[method] != "function") {
        continue;
      }
      wrapped[method] = (req: any, initReq: any) =>
        this.makeRequest(() => api[method](req, initReq));
    }
    return wrapped;
  }

  private static getOrInitiateRefresh(): Promise<void> {
    if (!this.refreshPromise) {
      this.refreshPromise = this.refreshTokenFn()?.finally(() => {
        // Set the promise back to null once the refresh operation is completed
        this.refreshPromise = null;
      });
    }
    return this.refreshPromise;
  }

  private static async makeRequest(fn: any): Promise<any> {
    // Wait for any ongoing refresh, if there is one, to complete
    if (this.refreshPromise) {
      await this.refreshPromise;
    }

    try {
      // Call the function directly
      return await fn();
    } catch (error) {
      // Check for a 401 status code on the HTTPError
      if (error.code === 401) {
        await this.getOrInitiateRefresh();
        // Try the request again
        return this.makeRequest(fn);
      }

      // Rethrow the error if it's not about the token
      throw error;
    }
  }
}

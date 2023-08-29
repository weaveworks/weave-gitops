import { TokenRefreshWrapper } from "../requests";

describe("ServiceWrapper", () => {
  const mockGetData = jest.fn();
  const mockRefreshToken = jest.fn(
    () => new Promise<void>((resolve) => setTimeout(resolve, 100))
  );
  const service = { getData: mockGetData };

  const wrappedService = TokenRefreshWrapper.wrap(service, mockRefreshToken);

  beforeEach(() => {
    mockGetData.mockReset();
    mockRefreshToken.mockClear();
  });

  test("it should refresh the token when a 401 error occurs and retry the request", async () => {
    mockGetData
      .mockRejectedValueOnce({ code: 401, message: "Unauthorized" })
      .mockResolvedValueOnce("Some Data");

    const data = await wrappedService.getData();

    expect(data).toBe("Some Data");
    expect(mockRefreshToken).toHaveBeenCalledTimes(1);
    expect(mockGetData).toHaveBeenCalledTimes(2);
  });

  test("it should not refresh the token if there is no 401 error", async () => {
    mockGetData.mockResolvedValue("Some Data");

    const data = await wrappedService.getData();

    expect(data).toBe("Some Data");
    expect(mockRefreshToken).toHaveBeenCalledTimes(0);
    expect(mockGetData).toHaveBeenCalledTimes(1);
  });

  test("it should refresh the token only once for concurrent requests", async () => {
    let tokenRefreshed = false;

    mockGetData.mockImplementation(() => {
      if (!tokenRefreshed) {
        return Promise.reject({ code: 401, message: "Unauthorized" });
      } else {
        return Promise.resolve("Some Data");
      }
    });

    mockRefreshToken.mockImplementation(() => {
      return new Promise<void>((resolve) => {
        setTimeout(() => {
          tokenRefreshed = true;
          resolve();
        }, 100);
      });
    });

    const [data1, data2] = await Promise.all([
      wrappedService.getData(),
      wrappedService.getData(),
    ]);

    expect(data1).toBe("Some Data");
    expect(data2).toBe("Some Data");
    expect(mockRefreshToken).toHaveBeenCalledTimes(1);
    expect(mockGetData).toHaveBeenCalledTimes(4);
  });
});

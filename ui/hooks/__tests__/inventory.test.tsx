import { ReadyStatusValue } from "../../components/KubeStatusIndicator";
import { FluxObject } from "../../lib/objects";
import { createCanaryCondition } from "../inventory";

describe("createCanaryCondition", () => {
  const falseCanary = {
    type: "Canary",
    conditions: [{ reason: "Failed" }],
  };
  const trueCanary = {
    type: "Canary",
    conditions: [{ reason: "Succeeded" }],
  };
  const unknownCanary = {
    type: "Canary",
    conditions: [{ reason: "Progressing" }],
  };
  it("creates a false condition if one or more canaries are failing", () => {
    expect(
      createCanaryCondition([
        falseCanary,
        trueCanary,
        unknownCanary,
      ] as FluxObject[]).status
    ).toEqual(ReadyStatusValue.False);
  });
  it("creates an unknown condition if no canaries are failing and one or more are unknown", () => {
    expect(
      createCanaryCondition([
        trueCanary,
        trueCanary,
        unknownCanary,
      ] as FluxObject[]).status
    ).toEqual(ReadyStatusValue.Unknown);
  });
  it("creates a true condition if all canaries succeeded", () => {
    expect(
      createCanaryCondition([
        trueCanary,
        trueCanary,
        trueCanary,
      ] as FluxObject[]).status
    ).toEqual(ReadyStatusValue.True);
  });
  it("creates a special condition if there are no canaries", () => {
    expect(createCanaryCondition([]).status).toEqual(ReadyStatusValue.None);
  });
});

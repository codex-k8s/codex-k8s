import { missionControlPrototypeModel } from "./fixtures";
import type { MissionControlPrototypeError, MissionControlPrototypeModel } from "./types";

type MissionControlPrototypeSource = {
  loadModel(): Promise<MissionControlPrototypeModel>;
};

export class MissionControlPrototypeSourceError extends Error {
  readonly uiError: MissionControlPrototypeError;

  constructor(messageKey: string, debugMessage?: string) {
    super(debugMessage || messageKey);
    this.name = "MissionControlPrototypeSourceError";
    this.uiError = {
      messageKey,
      debugMessage,
    };
  }
}

function cloneFixture<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

export const missionControlPrototypeSource: MissionControlPrototypeSource = {
  async loadModel() {
    if (missionControlPrototypeModel.projects.length === 0) {
      throw new MissionControlPrototypeSourceError(
        "pages.missionControlPrototype.errors.modelNotReady",
        "mission control prototype model is empty",
      );
    }
    return cloneFixture(missionControlPrototypeModel);
  },
};

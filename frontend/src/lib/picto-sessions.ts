import { Picto, type Room } from "../../bindings/changeme";

type RoomChoice = "create_room" | "join_room" | null;

export type Session = {
  roomChoice: RoomChoice;
  hasRoom: boolean;
  connected: boolean;
  isHost: boolean;
  websocket: WebSocket | null;
  room: Room | null;
};

export async function closeSession(session: Session) {
  if (session.isHost) {
    try {
      await Picto.StopServers();
    } catch (error) {
      console.error(error);
    }
  }

  console.log("Closing Session...");

  // Reset Session on cancel
  session.roomChoice = null;
  session.hasRoom = false;
  session.connected = false;
  session.isHost = false;
  session.websocket = null;
  session.room = null;
}

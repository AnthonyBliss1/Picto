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
  console.log("Closing Session...");
  session.websocket?.close();

  // avoid triping null check on close events once server shutsdown
  session.websocket = null;

  if (session.isHost) {
    try {
      console.log("Shutting down servers...");
      await Picto.StopServers();
    } catch (error) {
      console.error(error);
    }
  }

  // reset rest of session fields
  session.roomChoice = null;
  session.hasRoom = false;
  session.connected = false;
  session.isHost = false;
  session.room = null;
}

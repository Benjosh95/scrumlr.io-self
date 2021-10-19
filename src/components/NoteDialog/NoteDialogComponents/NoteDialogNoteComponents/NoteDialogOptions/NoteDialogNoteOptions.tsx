import Parse from "parse";
import {FC} from "react";
import {NoteDialogDeleteNoteButton} from "./NoteDialogDeleteNoteButton";
import {NoteDialogUnstackNoteButton} from "./NoteDialogUnstackNoteButton";
import "./NoteDialogNoteOptions.scss";

type NoteDialogNoteOptionsProps = {
  showUnstackButton: boolean;
  noteId?: string;
  authorId: string;
  onDeleteOfParent: () => void;
  onClose: () => void;
};

export const NoteDialogNoteOptions: FC<NoteDialogNoteOptionsProps> = (props: NoteDialogNoteOptionsProps) => {
  const showDeleteButton = props.authorId === Parse.User.current()?.id;
  return (
    <aside>
      <ul className="note-dialog__note-options">
        {showDeleteButton && (
          <li className="note-dialog__note-option">
            <NoteDialogDeleteNoteButton {...props} />
          </li>
        )}
        {props.showUnstackButton && (
          <li className="note-dialog__note-option">
            <NoteDialogUnstackNoteButton {...props} />
          </li>
        )}
      </ul>
    </aside>
  );
};
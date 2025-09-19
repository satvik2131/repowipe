import * as AlertDialog from "@radix-ui/react-alert-dialog";
import { Dispatch, SetStateAction } from "react";

export const ConfirmDialog = ({
  open,
  onOpenChange,
  onConfirm,
}: {
  open: boolean;
  onOpenChange: Dispatch<SetStateAction<boolean>>;
  onConfirm: () => void;
}) => {
  return (
    <AlertDialog.Root open={open} onOpenChange={onOpenChange}>
      <AlertDialog.Portal>
        {/* Fullscreen overlay */}
        <AlertDialog.Overlay className="fixed inset-0 bg-background/80" />

        {/* Centered content */}
        <AlertDialog.Content
          className="
            fixed top-1/2 left-1/2 
            -translate-x-1/2 -translate-y-1/2
            bg-background
            p-6 rounded-2xl shadow-lg 
            max-w-md w-full
            border border-border
          "
        >
          <AlertDialog.Title className="text-lg font-semibold text-foreground">
            Are you absolutely sure?
          </AlertDialog.Title>

          <AlertDialog.Description className="mt-2 text-sm text-muted-foreground">
            This action cannot be undone. This will permanently delete your
            Repositories.
          </AlertDialog.Description>

          <div className="flex gap-2 justify-end mt-4">
            <AlertDialog.Cancel asChild>
              <button className="px-4 py-2 rounded-lg bg-secondary text-secondary-foreground hover:bg-secondary/90">
                Cancel
              </button>
            </AlertDialog.Cancel>
            <AlertDialog.Action asChild>
              <button
                onClick={onConfirm}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700"
              >
                Yes, delete account
              </button>
            </AlertDialog.Action>
          </div>
        </AlertDialog.Content>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  );
};

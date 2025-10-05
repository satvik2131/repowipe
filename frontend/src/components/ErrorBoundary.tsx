import React from "react";
import * as AlertDialog from "@radix-ui/react-alert-dialog";

type ErrorBoundaryState = {
  hasError: boolean;
  error?: Error;
};

export class ErrorBoundary extends React.Component<
  React.PropsWithChildren,
  ErrorBoundaryState
> {
  constructor(props: React.PropsWithChildren) {
    super(props);
    this.state = { hasError: false, error: undefined };
  }

  static getDerivedStateFromError(error: Error) {
    // Trigger fallback UI
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error("Error caught by ErrorBoundary:", error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <AlertDialog.Root defaultOpen>
          <AlertDialog.Portal>
            <AlertDialog.Overlay className="fixed inset-0 bg-black/50" />
            <AlertDialog.Content className="fixed left-1/2 top-1/2 w-[90%] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl bg-white p-6 shadow-lg focus:outline-none">
              <AlertDialog.Title className="text-lg font-bold text-red-600">
                Oops! Something went wrong
              </AlertDialog.Title>
              <AlertDialog.Description className="mt-2 text-sm text-gray-600">
                {this.state.error?.message || "An unexpected error occurred."}
              </AlertDialog.Description>

              <div className="mt-4 flex justify-end gap-2">
                <AlertDialog.Cancel asChild>
                  <button
                    className="rounded-md border px-3 py-1 text-sm hover:bg-gray-100"
                    onClick={() => this.setState({ hasError: false })}
                  >
                    Dismiss
                  </button>
                </AlertDialog.Cancel>
                <AlertDialog.Action asChild>
                  <button
                    className="rounded-md bg-red-600 px-3 py-1 text-sm text-white hover:bg-red-700"
                    onClick={() => window.location.reload()}
                  >
                    Reload App
                  </button>
                </AlertDialog.Action>
              </div>
            </AlertDialog.Content>
          </AlertDialog.Portal>
        </AlertDialog.Root>
      );
    }

    return this.props.children;
  }
}

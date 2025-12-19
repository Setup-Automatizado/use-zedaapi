"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";

interface ErrorBoundaryProps {
	children: React.ReactNode;
	fallback?: (error: Error, reset: () => void) => React.ReactNode;
}

interface ErrorBoundaryState {
	hasError: boolean;
	error: Error | null;
}

export class ErrorBoundary extends React.Component<
	ErrorBoundaryProps,
	ErrorBoundaryState
> {
	constructor(props: ErrorBoundaryProps) {
		super(props);
		this.state = { hasError: false, error: null };
	}

	static getDerivedStateFromError(error: Error): ErrorBoundaryState {
		return { hasError: true, error };
	}

	componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
		console.error("Error caught by boundary:", error, errorInfo);
	}

	reset = () => {
		this.setState({ hasError: false, error: null });
	};

	render() {
		if (this.state.hasError && this.state.error) {
			if (this.props.fallback) {
				return this.props.fallback(this.state.error, this.reset);
			}

			return (
				<div className="flex min-h-[400px] flex-col items-center justify-center rounded-lg border border-destructive/50 bg-destructive/5 p-8 text-center">
					<div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
						<AlertTriangle className="h-8 w-8 text-destructive" />
					</div>
					<h3 className="mb-2 text-lg font-semibold">
						Something went wrong
					</h3>
					<p className="mb-6 max-w-md text-sm text-muted-foreground">
						{this.state.error.message ||
							"An unexpected error occurred"}
					</p>
					<Button onClick={this.reset}>Try again</Button>
				</div>
			);
		}

		return this.props.children;
	}
}

export function withErrorBoundary<P extends object>(
	Component: React.ComponentType<P>,
	fallback?: (error: Error, reset: () => void) => React.ReactNode,
) {
	return function WithErrorBoundary(props: P) {
		return (
			<ErrorBoundary fallback={fallback}>
				<Component {...props} />
			</ErrorBoundary>
		);
	};
}

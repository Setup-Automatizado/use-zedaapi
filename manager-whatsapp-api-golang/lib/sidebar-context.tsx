"use client";

import * as React from "react";

interface SidebarContextType {
	isCollapsed: boolean;
	setIsCollapsed: (collapsed: boolean) => void;
	toggle: () => void;
}

const SidebarContext = React.createContext<SidebarContextType | undefined>(
	undefined,
);

const STORAGE_KEY = "sidebar-collapsed";

export function SidebarProvider({ children }: { children: React.ReactNode }) {
	const [isCollapsed, setIsCollapsedState] = React.useState(false);
	const [isHydrated, setIsHydrated] = React.useState(false);

	// Load from localStorage on mount
	React.useEffect(() => {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === "true") {
			setIsCollapsedState(true);
		}
		setIsHydrated(true);
	}, []);

	const setIsCollapsed = React.useCallback((collapsed: boolean) => {
		setIsCollapsedState(collapsed);
		localStorage.setItem(STORAGE_KEY, String(collapsed));
	}, []);

	const toggle = React.useCallback(() => {
		setIsCollapsed(!isCollapsed);
	}, [isCollapsed, setIsCollapsed]);

	// Prevent hydration mismatch by not rendering children until hydrated
	if (!isHydrated) {
		return null;
	}

	return (
		<SidebarContext.Provider value={{ isCollapsed, setIsCollapsed, toggle }}>
			{children}
		</SidebarContext.Provider>
	);
}

export function useSidebar() {
	const context = React.useContext(SidebarContext);
	if (context === undefined) {
		throw new Error("useSidebar must be used within a SidebarProvider");
	}
	return context;
}

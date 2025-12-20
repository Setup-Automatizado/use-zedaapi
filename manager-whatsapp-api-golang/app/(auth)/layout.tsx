import type { Metadata } from "next";
import Image from "next/image";

export const metadata: Metadata = {
	title: "Login - WhatsApp Manager",
	description: "Access the WhatsApp instance management panel",
};

export default function AuthLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<div className="min-h-screen flex">
			{/* Left side - Branding */}
			<div className="hidden lg:flex lg:w-1/2 bg-primary/5 dark:bg-primary/10 relative overflow-hidden">
				<div className="absolute inset-0 bg-gradient-to-br from-primary/20 via-transparent to-transparent" />
				<div className="relative z-10 flex flex-col justify-center px-12 xl:px-20">
					<div className="space-y-6">
						<div className="flex items-center gap-3">
							<Image
								src="/android-chrome-96x96.png"
								alt="WhatsApp Manager"
								width={48}
								height={48}
								className="rounded-xl"
								priority
							/>
							<span className="text-2xl font-bold text-foreground">
								WhatsApp Manager
							</span>
						</div>
						<div className="space-y-4">
							<h1 className="text-4xl xl:text-5xl font-bold text-foreground leading-tight">
								Manage your WhatsApp instances
							</h1>
							<p className="text-lg text-muted-foreground max-w-md">
								Centralized panel to create, configure and
								monitor your WhatsApp connections simply and
								efficiently.
							</p>
						</div>
						<div className="flex flex-col gap-4 pt-4">
							<div className="flex items-center gap-3">
								<div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="2"
										strokeLinecap="round"
										strokeLinejoin="round"
										className="h-5 w-5 text-primary"
										aria-hidden="true"
									>
										<rect
											width="7"
											height="7"
											x="3"
											y="3"
											rx="1"
										/>
										<rect
											width="7"
											height="7"
											x="14"
											y="3"
											rx="1"
										/>
										<rect
											width="7"
											height="7"
											x="14"
											y="14"
											rx="1"
										/>
										<rect
											width="7"
											height="7"
											x="3"
											y="14"
											rx="1"
										/>
									</svg>
								</div>
								<div>
									<p className="font-medium text-foreground">
										QR Code & Phone Pairing
									</p>
									<p className="text-sm text-muted-foreground">
										Connect easily via QR or code
									</p>
								</div>
							</div>
							<div className="flex items-center gap-3">
								<div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="2"
										strokeLinecap="round"
										strokeLinejoin="round"
										className="h-5 w-5 text-primary"
										aria-hidden="true"
									>
										<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
										<circle cx="12" cy="12" r="3" />
									</svg>
								</div>
								<div>
									<p className="font-medium text-foreground">
										Configurable Webhooks
									</p>
									<p className="text-sm text-muted-foreground">
										7 event types available
									</p>
								</div>
							</div>
							<div className="flex items-center gap-3">
								<div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="2"
										strokeLinecap="round"
										strokeLinejoin="round"
										className="h-5 w-5 text-primary"
										aria-hidden="true"
									>
										<path d="M22 12h-4l-3 9L9 3l-3 9H2" />
									</svg>
								</div>
								<div>
									<p className="font-medium text-foreground">
										Real-Time Monitoring
									</p>
									<p className="text-sm text-muted-foreground">
										API status and health
									</p>
								</div>
							</div>
						</div>
					</div>
				</div>
				{/* Decorative elements */}
				<div className="absolute bottom-0 right-0 w-96 h-96 bg-primary/5 rounded-full blur-3xl -mb-48 -mr-48" />
				<div className="absolute top-0 left-0 w-64 h-64 bg-primary/5 rounded-full blur-3xl -mt-32 -ml-32" />
			</div>

			{/* Right side - Auth Form */}
			<div className="w-full lg:w-1/2 flex items-center justify-center p-6 sm:p-12">
				<div className="w-full max-w-md">{children}</div>
			</div>
		</div>
	);
}

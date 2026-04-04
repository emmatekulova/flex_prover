import { Verifier } from "@/components/verifier"

export default async function VerifyPage({
  params,
}: {
  params: Promise<{ hash: string }>
}) {
  const { hash } = await params

  return (
    <main className="min-h-screen bg-background flex flex-col items-center justify-center p-6">
      <Verifier initialHash={hash} />
    </main>
  )
}

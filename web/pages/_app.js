import dynamic from 'next/dynamic'
import '../styles/globals.css'

const MainNoSSR = dynamic(() => import('../components/main'), { ssr: false })

export default function JsonnetPlayground({ Component, pageProps }) {
  const Main = MainNoSSR
  return (<Main><Component /></Main>)
}
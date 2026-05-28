import { Outlet } from 'react-router-dom';
import Navbar from '../../components/Navbar/Navbar';
import Footer from '../../components/Footer/Footer';

/**
 * PublicLayout — wraps all pre-login (public) pages.
 * Every page rendered inside this layout automatically gets
 * the Navbar at top and Footer at bottom.
 *
 * React Router's <Outlet /> renders the current page's content
 * between the Navbar and Footer.
 */
export default function PublicLayout() {
  return (
    <>
      <Navbar />
      <main className="app-main">
        <Outlet />
      </main>
      <Footer />
    </>
  );
}

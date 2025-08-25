import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from '@/components/ui/navigation-menu.tsx';
import { Link } from '@tanstack/react-router';
import { LoginDialog } from '@/components/LoginDialog.tsx';
import { Button } from '@/components/ui/button.tsx';
import { useAuth } from '@/hooks/useAuth.tsx';
import UserDropdown from '@/components/UserDropdown.tsx';

export function Header() {
  const { login, isLoggingIn, logout, isAuthenticated, user } = useAuth();

  return (
    <header className="w-full border-b">
      <div className="container mx-auto px-6 py-4 flex items-center justify-between">
        <NavigationMenu className="mx-auto">
          <NavigationMenuList>
            <NavigationMenuItem>
              <NavigationMenuLink asChild>
                <Link to={'/'}>Home</Link>
              </NavigationMenuLink>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuLink asChild>
                <Link to={'/about'}>About</Link>
              </NavigationMenuLink>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>

        <div className={'w-[76px]'}>
          {isAuthenticated && user ? (
            <UserDropdown user={user} onLogout={logout} />
          ) : (
            <LoginDialog login={login} isLoggingIn={isLoggingIn}>
              <Button>Login</Button>
            </LoginDialog>
          )}
        </div>
      </div>
    </header>
  );
}

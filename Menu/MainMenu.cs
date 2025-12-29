using CustosAC.UI;
using CustosAC.WinAPI;

namespace CustosAC.Menu;

public static class MainMenu
{
    public static void Run()
    {
        while (true)
        {
            ConsoleUI.PrintHeader();
            ConsoleUI.PrintMenu("ГЛАВНОЕ МЕНЮ", new[]
            {
                "Ручная проверка",
                "Автоматическая проверка",
                "EXXXXXTRA"
            }, false);

            int choice = ConsoleUI.GetChoice(3);

            switch (choice)
            {
                case 0:
                    ConsoleUI.ClearScreen();
                    Console.WriteLine("\n\n");
                    Console.WriteLine($"  {ConsoleUI.ColorCyan}{ConsoleUI.ColorBold}╔════════════════════════════════════════════╗{ConsoleUI.ColorReset}");
                    Console.WriteLine($"  {ConsoleUI.ColorCyan}║                                            ║{ConsoleUI.ColorReset}");
                    Console.WriteLine($"  {ConsoleUI.ColorCyan}║{ConsoleUI.ColorReset}     {ConsoleUI.ColorYellow}+ Закрываем открытые процессы... +{ConsoleUI.ColorReset}    {ConsoleUI.ColorCyan}║{ConsoleUI.ColorReset}");
                    Console.WriteLine($"  {ConsoleUI.ColorCyan}║                                            ║{ConsoleUI.ColorReset}");
                    Console.WriteLine($"  {ConsoleUI.ColorCyan}{ConsoleUI.ColorBold}╚════════════════════════════════════════════╝{ConsoleUI.ColorReset}");
                    Thread.Sleep(800);
                    AdminHelper.KillAllProcesses();
                    ConsoleUI.PrintCleanupMessage();
                    return;
                case 1:
                    ManualMenu.Run();
                    break;
                case 2:
                    AutoMenu.Run();
                    break;
                case 3:
                    ExtraMenu.Run();
                    break;
            }
        }
    }
}

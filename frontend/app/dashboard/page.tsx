"use client"

import { SiteHeader } from "@/components/site-header";
import { TodoForm } from "@/components/todo-form";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useEffect, useState } from "react";

type Item = {
    id : string,
    name : string
}

export default function Dashboard() {

    const [items, setItems] = useState<Item[]>([]);

    useEffect(() => {
        // アイテムを取得する
        fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/users/userId/items`)
          .then((response) => response.json())
          .then((data) => setItems(data.items))
          .catch((error) => console.error('Error fetching items:', error));
      }, []);

    // アイテム追加関数
    function addItem(data: { name: string }) {

        const newItem: Item = {
            id: crypto.randomUUID(), // ユニークなIDを生成
            name: data.name
        }

        fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/users/userId/items`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(newItem),
          })
            .then((response) => response.json())
            .then((data) => {
              setItems([...items, data.item]);
            })
            .catch((error) => console.error('Error adding item:', error));

        setItems([...items, newItem]);
    }

    // アイテム削除関数
    function removeItem(itemId: string) {

        fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/users/userId/items/${itemId}`, {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
            }
          })
            .then((response) => response.status)
            .then((data) => {
                setItems(items.filter(item => item.id !== itemId));
            })
            .catch((error) => console.error('Error delete item:', error));
    }

    return (
        <div>
            <SiteHeader />
            <div className="p-6">
            <TodoForm onAddItem={addItem}/>
            <Table>
                <TableHeader>
                    <TableRow>
                    <TableHead className="w-[100px]">ID</TableHead>
                    <TableHead>道具名</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {items.map((item) => (
                        <TableRow key={item.id}>
                            <TableCell className="font-medium">{item.id}</TableCell>
                            <TableCell>{item.name}</TableCell>
                            <TableCell className="text-right"><Button onClick={() => removeItem(item.id)}>Close</Button></TableCell>
                        </TableRow>
                    ))}
                </TableBody>
                <TableFooter>
                    <TableRow>
                    <TableCell colSpan={4}>{items.length}のアイテムがあります。</TableCell>
                    </TableRow>
                </TableFooter>
            </Table>
            </div>
        </div>
    )
}